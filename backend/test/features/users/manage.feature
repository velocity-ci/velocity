Feature: Manage users
In order to authenticate teams
As an API user
I need to be able to manage users

  Background:
    Given I am authenticated

  Scenario: Create user
    When I create the following user:
      | username | testUser     |
      | password | testUser1234 |
    Then the response has status "201 Created"
    And the response has the following attributes:
      | attribute | type   | value    |
      | username  | string | testUser |

  Scenario: List users
    When I list the users
    Then the response has status "200 OK"
    And the response has the following attributes:
      | attribute        | type    | value |
      | total            | integer | 1     |
      | data[0].username | string  | admin |
