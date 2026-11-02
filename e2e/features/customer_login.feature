# language: en
Feature: Customer login through the login lambda
  POST /auth/customer-login is served by the lambda (through the adapter). Its token is accepted by the services.

  Scenario: A known CPF receives a token that the services accept
    Given the gateway is ready
    And I am logged in as "attendant"
    And I am logged in as "mechanic"
    When I create a service order for customer A with the oil change work
    And the attendant marks the vehicle as received
    And the mechanic starts the diagnosis
    And the mechanic adds 2 units of the oil filter
    And the mechanic sends the budget
    And the budget is generated and the order awaits approval
    When the customer logs in with CPF "52998224725"
    Then the lambda answers 200
    And the response contains a token
    When the customer accepts the budget
    Then the service order status is "IN_PROGRESS"

  Scenario: An unknown CPF is refused
    Given the gateway is ready
    When the customer logs in with CPF "00000000000"
    Then the lambda answers 404
