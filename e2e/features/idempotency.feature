# language: en
Feature: Repeated requests do not duplicate effects
  Repeating an accepted decision or a payment confirmation answers 409 and changes nothing.

  Background:
    Given the gateway is ready
    And I am logged in as "attendant"
    And I am logged in as "mechanic"
    And I am logged in as customer A through the login lambda

  Scenario: Accept and confirm sent twice
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
    When I send the accept request again
    Then the gateway answers 409
    When I send the payment confirmation again
    Then the gateway answers 409
    And the execution queue lists exactly one task for the service order
    And the service order status is "IN_PROGRESS"
