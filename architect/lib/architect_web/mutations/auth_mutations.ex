defmodule ArchitectWeb.Mutations.AuthMutations do
  use Absinthe.Schema.Notation
  alias Architect.Accounts
  alias Kronky.ValidationMessage

  object :auth_mutations do
    @desc "Sign in"
    field :sign_in, :session_payload do
      arg(:username, non_null(:string))
      arg(:password, non_null(:string))

      resolve(fn %{username: username, password: password}, %{context: _context} ->
        with {:ok, user} <- Accounts.authenticate(username, password),
             {:ok, token, _} <- Accounts.encode_and_sign(user) do
          {:ok, %{token: token, username: username}}
        else
          {:error, :invalid_credentials} ->
            {:error, "Incorrect username or password"}
        end
      end)
    end
  end
end
