package e2e

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

// Fixed identifiers of the seed (docs/fase4/contracts/resources.md). Changing them here means changing the seed too.
const (
	customerAID     = "44444444-4444-4444-8444-000000000002" // customer A (Joao Silva), user ...001
	vehicleAID      = "44444444-4444-4444-8444-000000000003"
	customerACPF    = "52998224725"
	workOilChange   = "11111111-1111-4111-8111-000000000001" // "Troca de Oleo"
	supplyOilFilter = "22222222-2222-4222-8222-000000000001" // "Filtro de Oleo", stock 50 in the seed
	readyTimeout    = 3 * time.Minute
	stepTimeout     = 20 * time.Second
	pollInterval    = 500 * time.Millisecond
)

// staffCredentials are the seed users of the customer service (SEED_ENABLED=true). They are local development values.
var staffCredentials = map[string][2]string{
	"attendant": {"attendant@autorepairshop.com", "attendant123"},
	"mechanic":  {"mechanic@autorepairshop.com", "mechanic123"},
}

// world is the state of one scenario. godog calls initializeScenario once per scenario, so nothing leaks between them.
type world struct {
	c             *httpClient
	tokens        map[string]string // "attendant", "mechanic", "customer"
	soID          string
	works         []string
	paymentID     string
	stockBaseline int
	last          *reply
}

func initializeScenario(sc *godog.ScenarioContext) {
	w := &world{c: newHTTPClient(), tokens: map[string]string{}}

	// Background
	sc.Step(`^the gateway is ready$`, w.gatewayReady)
	sc.Step(`^I am logged in as "(attendant|mechanic)"$`, w.staffLogin)
	sc.Step(`^I am logged in as customer A through the login lambda$`, w.loginCustomerA)
	sc.Step(`^I record the stock of the oil filter$`, w.recordStock)

	// Service order creation and the budget
	sc.Step(`^I create a service order for customer A with the oil change work$`, w.createServiceOrder)
	sc.Step(`^the attendant marks the vehicle as received$`, w.markReceived)
	sc.Step(`^the mechanic starts the diagnosis$`, w.startDiagnosis)
	sc.Step(`^the mechanic adds 2 units of the oil filter$`, w.addOilFilter)
	sc.Step(`^the mechanic sends the budget$`, w.sendBudget)
	sc.Step(`^the budget is generated and the order awaits approval$`, w.budgetReady)
	sc.Step(`^the customer accepts the budget$`, w.customerAccepts)
	sc.Step(`^the customer rejects the budget$`, w.customerRejects)

	// Payment (billing runs in manual mode: the attendant decides)
	sc.Step(`^the payment is pending$`, w.paymentPending)
	sc.Step(`^the attendant approves the payment$`, w.approvePayment)
	sc.Step(`^the attendant rejects the payment$`, w.rejectPayment)

	// Execution
	sc.Step(`^the execution queue lists the service order$`, w.queueListsOrder)
	sc.Step(`^the mechanic completes every work of the order$`, w.completeWorks)
	sc.Step(`^the mechanic cancels every work of the order$`, w.cancelWorks)
	sc.Step(`^the mechanic finishes the service order$`, w.finishOrder)
	sc.Step(`^the attendant delivers the vehicle$`, w.deliver)

	// Assertions on the outcome
	sc.Step(`^the service order status is "([A-Z_]+)"$`, w.statusIs)
	sc.Step(`^the payment status is "([A-Z]+)"$`, w.paymentStatusIs)
	sc.Step(`^the stock of the oil filter is back to its initial value$`, w.stockRestored)
	sc.Step(`^the history lists the transitions with work transitions$`, w.historyLists)
	sc.Step(`^the average execution time report includes the work$`, w.reportIncludesWork)

	// Customer login through the lambda
	sc.Step(`^the customer logs in with CPF "([0-9]+)"$`, w.customerLoginWithCPF)
	sc.Step(`^the lambda answers (\d+)$`, w.lastStatusIs)
	sc.Step(`^the response contains a token$`, w.responseHasToken)

	// Security
	sc.Step(`^I request the internal customer endpoint of customer A without a token$`, w.internalRequest)
	sc.Step(`^I list the service orders without a token$`, w.listWithoutToken)
	sc.Step(`^the gateway answers (\d+)$`, w.lastStatusIs)
	sc.Step(`^the body code is (\d+)$`, w.bodyCodeIs)

	// Idempotency
	sc.Step(`^I send the accept request again$`, w.acceptAgain)
	sc.Step(`^I send the payment confirmation again$`, w.confirmAgain)
	sc.Step(`^the execution queue lists exactly one task for the service order$`, w.queueHasOneTask)
}

// ---- helpers ------------------------------------------------------------------------------------------------------

func (w *world) soPath(suffix string) string {
	return "/v1/service-order/" + w.soID + suffix
}

// do sends a request as the given role ("" = anonymous) and remembers the reply.
func (w *world) do(method, path, role string, body any) error {
	r, err := w.c.send(method, path, w.tokens[role], body)
	if err != nil {
		return err
	}
	w.last = r
	return nil
}

// expectLast fails unless the last reply has the wanted status.
func (w *world) expectLast(want int) error {
	if w.last == nil {
		return fmt.Errorf("no reply yet")
	}
	if w.last.status != want {
		return fmt.Errorf("status %d, want %d", w.last.status, want)
	}
	return nil
}

// expectDo sends a request and checks the status it answers.
func (w *world) expectDo(method, path, role string, body any, want int) error {
	if err := w.do(method, path, role, body); err != nil {
		return err
	}
	return w.expectLast(want)
}

// statusOfOrder reads the public status endpoint (no token needed).
func (w *world) statusOfOrder() (string, error) {
	r, err := w.c.send("GET", w.soPath("/status"), "", nil)
	if err != nil {
		return "", err
	}
	if r.status != 200 {
		return "", fmt.Errorf("status endpoint answered %d", r.status)
	}
	var out struct {
		Status string `json:"status"`
	}
	if err := r.decode(&out); err != nil {
		return "", err
	}
	return out.Status, nil
}

// ---- background ----------------------------------------------------------------------------------------------------

func (w *world) gatewayReady() error {
	return poll(readyTimeout, time.Second, func() error {
		r, err := w.c.send("GET", "/ready", "", nil)
		if err != nil {
			return err
		}
		if r.status != 200 {
			return fmt.Errorf("ready answered %d", r.status)
		}
		return nil
	})
}

func (w *world) staffLogin(role string) error {
	cred := staffCredentials[role]
	if err := w.do("POST", "/v1/auth/login", "", map[string]string{"email": cred[0], "password": cred[1]}); err != nil {
		return err
	}
	if err := w.expectLast(200); err != nil {
		return fmt.Errorf("login as %s: %w", role, err)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := w.last.decode(&out); err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("login as %s returned no token", role)
	}
	w.tokens[role] = out.Token
	return nil
}

// customerLoginWithCPF calls POST /auth/customer-login. A 200 stores the token as the customer's.
func (w *world) customerLoginWithCPF(cpf string) error {
	if err := w.do("POST", "/auth/customer-login", "", map[string]string{"cpf": cpf}); err != nil {
		return err
	}
	if w.last.status != 200 {
		return nil
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := w.last.decode(&out); err != nil {
		return err
	}
	w.tokens["customer"] = out.Token
	return nil
}

func (w *world) loginCustomerA() error {
	if err := w.customerLoginWithCPF(customerACPF); err != nil {
		return err
	}
	if err := w.expectLast(200); err != nil {
		return fmt.Errorf("customer A login: %w", err)
	}
	return nil
}

func (w *world) recordStock() error {
	stock, err := w.stockOf(supplyOilFilter)
	if err != nil {
		return err
	}
	w.stockBaseline = stock
	return nil
}

func (w *world) stockOf(supplyID string) (int, error) {
	if err := w.do("GET", "/v1/supplies?pageSize=100", "attendant", nil); err != nil {
		return 0, err
	}
	if err := w.expectLast(200); err != nil {
		return 0, fmt.Errorf("list supplies: %w", err)
	}
	var out struct {
		Items []struct {
			ID    string `json:"id"`
			Stock int    `json:"stock_quantity"`
		} `json:"items"`
	}
	if err := w.last.decode(&out); err != nil {
		return 0, err
	}
	for _, it := range out.Items {
		if it.ID == supplyID {
			return it.Stock, nil
		}
	}
	return 0, fmt.Errorf("supply %s not in the list", supplyID)
}

// ---- service order ---------------------------------------------------------------------------------------------------

func (w *world) createServiceOrder() error {
	body := map[string]any{
		"client":  customerAID,
		"vehicle": vehicleAID,
		"works":   []string{workOilChange},
	}
	if err := w.do("POST", "/v1/service-order", "attendant", body); err != nil {
		return err
	}
	if err := w.expectLast(201); err != nil {
		return fmt.Errorf("create service order: %w", err)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := w.last.decode(&out); err != nil {
		return err
	}
	if out.ID == "" {
		return fmt.Errorf("create service order: no id in the reply")
	}
	w.soID = out.ID
	w.works = []string{workOilChange}
	return nil
}

func (w *world) statusIs(want string) error {
	return poll(stepTimeout, pollInterval, func() error {
		got, err := w.statusOfOrder()
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("status %s, want %s", got, want)
		}
		return nil
	})
}

func (w *world) markReceived() error {
	return w.expectDo("PUT", w.soPath("/received"), "attendant", nil, 204)
}

func (w *world) startDiagnosis() error {
	return w.expectDo("PUT", w.soPath("/start-diagnosis"), "mechanic", nil, 204)
}

func (w *world) addOilFilter() error {
	body := map[string]any{"supplies": []map[string]any{{"id": supplyOilFilter, "amount": 2}}}
	return w.expectDo("POST", w.soPath("/supplies"), "mechanic", body, 204)
}

func (w *world) sendBudget() error {
	return w.expectDo("PUT", w.soPath("/send"), "mechanic", nil, 204)
}

// budgetReady waits until billing has generated the budget and the order waits for the customer.
func (w *world) budgetReady() error {
	err := poll(stepTimeout, pollInterval, func() error {
		if err := w.do("GET", "/v1/budgets/"+w.soID, "attendant", nil); err != nil {
			return err
		}
		if w.last.status != 200 {
			return fmt.Errorf("budget answered %d", w.last.status)
		}
		var out struct {
			Status string `json:"status"`
		}
		if err := w.last.decode(&out); err != nil {
			return err
		}
		if out.Status != "GENERATED" {
			return fmt.Errorf("budget status %s", out.Status)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return w.statusIs("AWAITING_APPROVAL")
}

func (w *world) customerAccepts() error {
	return w.expectDo("PUT", w.soPath("/accept"), "customer", nil, 204)
}

func (w *world) customerRejects() error {
	return w.expectDo("PUT", w.soPath("/reject"), "customer", nil, 204)
}

func (w *world) acceptAgain() error {
	return w.do("PUT", w.soPath("/accept"), "customer", nil)
}

// ---- payment -------------------------------------------------------------------------------------------------------

// currentPayment reads the most recent payment of the order (attendant).
func (w *world) currentPayment() (id, status string, err error) {
	if err := w.do("GET", "/v1/payments/"+w.soID, "attendant", nil); err != nil {
		return "", "", err
	}
	if err := w.expectLast(200); err != nil {
		return "", "", fmt.Errorf("get payment: %w", err)
	}
	var out struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := w.last.decode(&out); err != nil {
		return "", "", err
	}
	return out.ID, out.Status, nil
}

func (w *world) paymentPending() error {
	return poll(stepTimeout, pollInterval, func() error {
		id, status, err := w.currentPayment()
		if err != nil {
			return err
		}
		if status != "PENDING" {
			return fmt.Errorf("payment status %s", status)
		}
		w.paymentID = id
		return nil
	})
}

func (w *world) approvePayment() error {
	return w.expectDo("POST", "/v1/payments/"+w.paymentID+"/confirm", "attendant", nil, 204)
}

func (w *world) rejectPayment() error {
	return w.expectDo("POST", "/v1/payments/"+w.paymentID+"/reject", "attendant", nil, 204)
}

func (w *world) paymentStatusIs(want string) error {
	return poll(stepTimeout, pollInterval, func() error {
		_, status, err := w.currentPayment()
		if err != nil {
			return err
		}
		if status != want {
			return fmt.Errorf("payment status %s, want %s", status, want)
		}
		return nil
	})
}

// ---- execution -----------------------------------------------------------------------------------------------------

// queueItems returns the active execution tasks of the service order.
func (w *world) queueItems() (int, error) {
	if err := w.do("GET", "/v1/execution/queue", "mechanic", nil); err != nil {
		return 0, err
	}
	if err := w.expectLast(200); err != nil {
		return 0, fmt.Errorf("execution queue: %w", err)
	}
	var out struct {
		Items []struct {
			ServiceOrderID string `json:"serviceOrderId"`
		} `json:"items"`
	}
	if err := w.last.decode(&out); err != nil {
		return 0, err
	}
	n := 0
	for _, it := range out.Items {
		if it.ServiceOrderID == w.soID {
			n++
		}
	}
	return n, nil
}

func (w *world) queueListsOrder() error {
	return poll(stepTimeout, pollInterval, func() error {
		n, err := w.queueItems()
		if err != nil {
			return err
		}
		if n < 1 {
			return fmt.Errorf("order not in the queue yet")
		}
		return nil
	})
}

func (w *world) queueHasOneTask() error {
	n, err := w.queueItems()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("queue has %d tasks for the order, want 1", n)
	}
	return nil
}

// completeWorks advances every work twice: AWAITING_START -> IN_PROGRESS -> COMPLETED.
func (w *world) completeWorks() error {
	for _, work := range w.works {
		for i := 0; i < 2; i++ {
			if err := w.expectDo("PUT", w.soPath("/work/"+work+"/next"), "mechanic", nil, 204); err != nil {
				return fmt.Errorf("advance work %s: %w", work, err)
			}
		}
	}
	return nil
}

func (w *world) cancelWorks() error {
	for _, work := range w.works {
		if err := w.expectDo("PUT", w.soPath("/work/"+work+"/cancel"), "mechanic", nil, 204); err != nil {
			return fmt.Errorf("cancel work %s: %w", work, err)
		}
	}
	return nil
}

func (w *world) finishOrder() error {
	return w.expectDo("PUT", w.soPath("/finish"), "mechanic", nil, 204)
}

func (w *world) deliver() error {
	return w.expectDo("PUT", w.soPath("/deliver"), "attendant", nil, 204)
}

// ---- outcome -------------------------------------------------------------------------------------------------------

func (w *world) stockRestored() error {
	return poll(stepTimeout, pollInterval, func() error {
		stock, err := w.stockOf(supplyOilFilter)
		if err != nil {
			return err
		}
		if stock != w.stockBaseline {
			return fmt.Errorf("stock %d, want %d", stock, w.stockBaseline)
		}
		return nil
	})
}

func (w *world) historyLists() error {
	return poll(stepTimeout, pollInterval, func() error {
		if err := w.do("GET", w.soPath("/history"), "attendant", nil); err != nil {
			return err
		}
		if err := w.expectLast(200); err != nil {
			return fmt.Errorf("history: %w", err)
		}
		var entries []struct {
			NewStatus       string `json:"newStatus"`
			WorkTransitions []struct {
				WorkID string `json:"workId"`
				Status []struct {
					NewStatus string `json:"newStatus"`
				} `json:"status"`
			} `json:"workTransitions"`
		}
		if err := w.last.decode(&entries); err != nil {
			return err
		}
		if len(entries) == 0 || entries[len(entries)-1].NewStatus != "DELIVERED" {
			return fmt.Errorf("history does not end in DELIVERED yet")
		}
		for _, e := range entries {
			if e.NewStatus != "IN_PROGRESS" {
				continue
			}
			for _, wt := range e.WorkTransitions {
				if wt.WorkID == workOilChange && len(wt.Status) >= 2 {
					return nil
				}
			}
		}
		return fmt.Errorf("no IN_PROGRESS entry with the work transitions of %s", workOilChange)
	})
}

func (w *world) reportIncludesWork() error {
	return poll(stepTimeout, pollInterval, func() error {
		if err := w.do("GET", "/v1/reports/average-execution-time", "attendant", nil); err != nil {
			return err
		}
		if err := w.expectLast(200); err != nil {
			return fmt.Errorf("report: %w", err)
		}
		var rows []struct {
			WorkID string `json:"workId"`
		}
		if err := w.last.decode(&rows); err != nil {
			return err
		}
		for _, r := range rows {
			if r.WorkID == workOilChange {
				return nil
			}
		}
		return fmt.Errorf("report does not include %s yet", workOilChange)
	})
}

// ---- lambda, security and idempotency -----------------------------------------------------------------------------

func (w *world) lastStatusIs(want string) error {
	n, err := strconv.Atoi(want)
	if err != nil {
		return err
	}
	return w.expectLast(n)
}

func (w *world) responseHasToken() error {
	if w.last == nil {
		return fmt.Errorf("no reply yet")
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := w.last.decode(&out); err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("reply without token")
	}
	return nil
}

func (w *world) internalRequest() error {
	return w.do("GET", "/internal/customers/"+customerAID, "", nil)
}

func (w *world) listWithoutToken() error {
	return w.do("GET", "/v1/service-order", "", nil)
}

func (w *world) bodyCodeIs(want string) error {
	n, err := strconv.Atoi(want)
	if err != nil {
		return err
	}
	if w.last == nil {
		return fmt.Errorf("no reply yet")
	}
	var out struct {
		Code int `json:"code"`
	}
	if err := w.last.decode(&out); err != nil {
		return err
	}
	if out.Code != n {
		return fmt.Errorf("body code %d, want %d", out.Code, n)
	}
	return nil
}

func (w *world) confirmAgain() error {
	return w.do("POST", "/v1/payments/"+w.paymentID+"/confirm", "attendant", nil)
}
