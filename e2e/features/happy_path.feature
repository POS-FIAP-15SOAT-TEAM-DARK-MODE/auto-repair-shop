# language: en
Feature: Service order happy path
  A customer's service order goes from reception to delivery through the gateway.

  Background:
    Given the gateway is ready
    And I am logged in as "attendant"
    And I am logged in as "mechanic"
    And I am logged in as customer A through the login lambda

  Scenario: Order is received, approved, paid, executed, delivered and recorded in the history
    When I create a service order for customer A with the oil change work
    Then the service order status is "NEW"
    When the attendant marks the vehicle as received
    Then the service order status is "RECEIVED"
    When the mechanic starts the diagnosis
    Then the service order status is "IN_DIAGNOSIS"
    When the mechanic adds 2 units of the oil filter
    And the mechanic sends the budget
    Then the budget is generated and the order awaits approval
    When the customer accepts the budget
    Then the service order status is "IN_PROGRESS"
    And the payment is pending
    When the attendant approves the payment
    Then the payment status is "APPROVED"
    And the execution queue lists the service order
    When the mechanic completes every work of the order
    And the mechanic finishes the service order
    Then the service order status is "COMPLETED"
    When the attendant delivers the vehicle
    Then the service order status is "DELIVERED"
    And the history lists the transitions with work transitions
    And the average execution time report includes the work
