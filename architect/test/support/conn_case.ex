defmodule ArchitectWeb.ConnCase do
  @moduledoc """
  This module defines the test case to be used by
  tests that require setting up a connection.

  Such tests rely on `Phoenix.ConnTest` and also
  import other functionality to make it easier
  to build common data structures and query the data layer.

  Finally, if the test case interacts with the database,
  it cannot be async. For this reason, every test runs
  inside a transaction which is reset at the beginning
  of the test unless the test case is marked as async.
  """

  use ExUnit.CaseTemplate
  alias Architect.Accounts

  using do
    quote do
      @test_user_params %{username: "test_user", password: "test_password"}

      # Import conveniences for testing with connections
      use Phoenix.ConnTest
      alias ArchitectWeb.Router.Helpers, as: Routes

      # The default endpoint for testing
      @endpoint ArchitectWeb.Endpoint

      def graphql_request(query) do
        with {:ok, user} <- Accounts.create_user(@test_user_params),
             {:ok, token, _} <- Accounts.encode_and_sign(user) do
          build_conn()
          |> put_req_header("authorization", "Bearer " <> token)
          |> post("/v2", %{query: query})
          |> json_response(200)
        else
          err ->
            IO.inspect(err)
            flunk("Failed authenticating")
        end
      end

      def unauthorized_graphql_request(query) do
        build_conn()
        |> post("/v2", %{query: query})
        |> json_response(200)
      end

      def expect_success!(response) do
        case response do
          %{"data" => data} ->
            data

          _ ->
            flunk("graphql response failed: #{inspect(response)}}")
        end
      end

      def expect_failure!(response) do
        case response do
          %{"errors" => [_e | _] = data} ->
            data

          _ ->
            flunk("graphql response succeeded: #{inspect(response)}}")
        end
      end
    end
  end

  setup tags do
    :ok = Ecto.Adapters.SQL.Sandbox.checkout(Architect.Repo)

    unless tags[:async] do
      Ecto.Adapters.SQL.Sandbox.mode(Architect.Repo, {:shared, self()})
    end

    {:ok, conn: Phoenix.ConnTest.build_conn()}
  end
end
