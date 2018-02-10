Feature: Authenticate users
In order to use the API
As an API user
I need to be able to authenticate


  Scenario: Valid credentials
    Given the following users exist:
      | username | password        |
      | admin    | testPassword123 |
    When I authenticate with the following credentials:
      | username | admin           |
      | password | testPassword123 |
    Then the response has status "201 Created"
    And the response has the following attributes:
      | attribute | type      | value |
      | username  | string    | admin |
      | token     | string    | *any  |
      | expires   | timestamp | *48h  |

  Scenario: Invalid password
    Given the following users exist:
      | username | password        |
      | admin    | testPassword123 |
    When I authenticate with the following credentials:
      | username | admin           |
      | password | testPassword124 |
    Then the response has status "401 Unauthorized"

  Scenario: Invalid username
    Given the following users exist:
      | username | password        |
      | admin    | testPassword123 |
    When I authenticate with the following credentials:
      | username | admim           |
      | password | testPassword123 |
    Then the response has status "401 Unauthorized"
