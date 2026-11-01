# language: en
Feature: Execution failure
  If every work is cancelled, the order is cancelled, the payment is refunded and the stock returns.

  Background:
    Given the gateway is ready
    And I am logged in as "attendant"
    And I am logged in as "mechanic"
    And I am logged in as customer A through the login lambda
    And I record the stock of the oil filter

  Scenario: Mechanic cancels all works after the payment is approved
    When I create a service order for customer A with the oil change work
    And the attendant marks the vehicle as received
    And the mechanic starts the diagnosis
    And the mechanic adds 2 units of the oil filter
    And the mechanic sends the budget
    And the budget is generated and the order awaits approval
    And the customer accepts the budget
    And the payment is pending
    And the attendant approves the payment
    And the execution queue lists the service order
    When the mechanic cancels every work of the order
    And the mechanic finishes the service order
    Then the service order status is "CANCELLED"
    And the payment status is "REFUNDED"
    And the stock of the oil filter is back to its initial value
