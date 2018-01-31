Feature: Create projects
In order to run CI against my code
As an API user
I need to be able to create a project

  Background:
    Given I am authenticated

  Scenario: Add a valid project with https
    When I create the following project:
      | attribute  | value                                         |
      | name       | Velocity public https                         |
      | repository | http://gogs:3000/velocity/velocity_public.git |
    Then the response has status "201 Created"
    And the response has the following attributes:
      | attribute | type   | value                 |
      | id        | string | *any                  |
      | name      | string | Velocity public https |