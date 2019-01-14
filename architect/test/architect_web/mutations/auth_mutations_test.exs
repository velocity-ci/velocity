defmodule ArchitectWeb.Mutations.AuthMutationsTest do
  alias Architect.Accounts
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper

  @users_attrs [
    %{username: "eddy", password: "password"}
  ]

  setup do
    users = for user_attrs <- @users_attrs, do: Accounts.create_user(user_attrs)

    [users: users]
  end

  test "User gets a JWT on successful login", %{users: [user|_]} do
    mutation =
      """
      mutation{
        signIn(
          username: eddy
          password: password
        ){
          successful
          messages {
            field
            message
            code
          }
          result {
            token,
            username
          }
        }
      }
      """
  end

end
