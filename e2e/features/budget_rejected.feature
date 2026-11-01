# language: en
Feature: Budget rejected by the customer
  If the customer rejects the budget, the order is rejected and the reserved stock returns.

  Background:
    Given the gateway is ready
    And I am logged in as "attendant"
    And I am logged in as "mechanic"
    And I am logged in as customer A through the login lambda
    And I record the stock of the oil filter

  Scenario: Customer rejects the budget
    When I create a service order for customer A with the oil change work
    And the attendant marks the vehicle as received
    And the mechanic starts the diagnosis
    And the mechanic adds 2 units of the oil filter
    And the mechanic sends the budget
    And the budget is generated and the order awaits approval
    When the customer rejects the budget
    Then the service order status is "REJECTED"
    And the stock of the oil filter is back to its initial value
