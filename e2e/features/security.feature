# language: en
Feature: Route security
  The gateway never forwards /internal/*, and the services reject calls without a token.

  Background:
    Given the gateway is ready

  Scenario: Internal endpoints are not reachable through the gateway
    When I request the internal customer endpoint of customer A without a token
    Then the gateway answers 404

  Scenario: A protected route without a token is refused by the service
    When I list the service orders without a token
    Then the gateway answers 401
    And the body code is 401
