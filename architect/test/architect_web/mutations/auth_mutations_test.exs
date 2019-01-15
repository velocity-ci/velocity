defmodule ArchitectWeb.Mutations.AuthMutationsTest do
  alias Architect.Accounts
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper
  alias Kronky.ValidationMessage

  @valid_password "valid_password"
  @invalid_password "invalid_password"

  @users_attrs [
    %{username: "eddy", password: @valid_password}
  ]

  @fields %{
    token: :string,
    username: :string
  }

  setup do
    users =
      for user_attrs <- @users_attrs do
        {:ok, user} = Accounts.create_user(user_attrs)
        user
      end

    [users: users]
  end

  describe "signIn" do
    test "Success", %{users: [user | _]} do
      mutation = "
      mutation {
        signIn(username: \"#{user.username}\", password: \"#{@valid_password}\") {
          result {
            token,
            username
          },
          successful
        }
      }
    "

      %{"signIn" => actual} =
        unauthorized_graphql_request(mutation)
        |> expect_success!()

      token = get_in(actual, ["result", "token"])

      expected = %{username: "eddy", token: token}

      assert_mutation_success(expected, actual, @fields)
    end

    test "Failure - No user with supplied username", %{} do
      mutation = "
        mutation {
          signIn(username: \"does-not-exist\", password: \"does-not-exist-password\") {
            result {
              token,
              username
            },
            successful,
            messages {
              field,
              code,
              message
            }
          }
        }
        "

      %{"signIn" => actual} =
        unauthorized_graphql_request(mutation)
        |> expect_success!()

      expected = %ValidationMessage{
        code: :invalid_credentials
      }

      assert_mutation_failure([expected], actual, [:code])
    end

    test "Failure - wrong password", %{users: [user | _]} do
      mutation = "
        mutation {
          signIn(username: \"#{user.username}\", password: \"#{@invalid_password}\") {
            result {
              token,
              username
            },
            successful,
            messages {
              field,
              code,
              message
            }
          }
        }
        "

      %{"signIn" => actual} =
        unauthorized_graphql_request(mutation)
        |> expect_success!()

      expected = %ValidationMessage{
        code: :invalid_credentials
      }

      assert_mutation_failure([expected], actual, [:code])
    end
  end
end
