defmodule ArchitectWeb.Mutations.AuthMutations do
  use Absinthe.Schema.Notation

  import ArchitectWeb.Helpers.ValidationMessageHelpers

  alias ArchitectWeb.Schema.Middleware
  alias ArchitectWeb.Email
  alias ArchitectWeb.{Accounts, Confirmations, Mailer}
  alias Architect.Users.Guardian

  object :auth_mutations do
    @desc "Sign in"
    field :sign_in, :session_payload do
      arg(:username, :string)
      arg(:password, :string)

      resolve(fn args, %{context: context} ->
        with {:ok, user} <- Accounts.authenticate(args[:username], args[:password]),
             {:ok, token, _} <- Accounts.encode_and_sign(user) do
          {:ok, %{token: token}}
        else
          _ ->
            {:error, generic_message("Email ou mot de passe invalide.")}
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
