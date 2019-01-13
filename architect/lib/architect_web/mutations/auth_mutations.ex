defmodule ArchitectWeb.Mutations.AuthMutations do
  use Absinthe.Schema.Notation

  import ArchitectWeb.Helpers.ValidationMessageHelpers

  alias ArchitectWeb.Schema.Middleware
  alias ArchitectWeb.Email
  alias Architect.Accounts

  object :auth_mutations do
    @desc "Sign in"
    field :sign_in, non_null(:session_payload) do
      arg(:username, non_null(:string))
      arg(:password, non_null(:string))

      resolve(fn %{username: username, password: password}, %{context: context} ->
        with {:ok, user} <- Accounts.authenticate(username, password),
             {:ok, token, _} <- Accounts.encode_and_sign(user) do
          {:ok, %{token: token, username: username}}
        else
          _ ->
            {:error, "Invalid credentials"}
        end
      end)
    end

    #    @desc "Revoke token"
    #    field :revoke_token, :boolean do
    #      middleware(Middleware.Authorize)
    #
    #      resolve(fn _, %{context: context} ->
    #        context[:current_user] |> Accounts.revoke_access_token()
    #        {:ok, true}
    #      end)
    #    end
  end
end
