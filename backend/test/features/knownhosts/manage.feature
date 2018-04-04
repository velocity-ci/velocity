Feature: Manage known hosts
In order to use Git SSH
As an API user
I need to be able to manage trusted hosts

  Background:
    Given I am authenticated

  Scenario: Add a trusted host
    When I create the following known host:
      """
      github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
      """
    Then the response has status "201 Created"
    And the response has the following attributes:
      | attribute | type   | value                                              |
      | id        | string | *any                                               |
      | hosts[0]  | string | github.com                                         |
      | comment   | string |                                                    |
      | sha256    | string | SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8 |
      | md5       | string | 16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48    |

  Scenario: List trusted hosts
    Given the following known host exists:
      """
      github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
      """
    When I list the known hosts
    Then the response has status "200 OK"
    And the response has the following attributes:
      | attribute        | type    | value                                              |
      | total            | integer | 1                                                  |
      | data[0].id       | string  | *any                                               |
      | data[0].hosts[0] | string  | github.com                                         |
      | data[0].comment  | string  |                                                    |
      | data[0].sha256   | string  | SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8 |
      | data[0].md5      | string  | 16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48    |
