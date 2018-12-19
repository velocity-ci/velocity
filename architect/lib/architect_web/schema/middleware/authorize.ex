defmodule ArchitectWeb.Schema.Middleware.Authorize do
  @behaviour Absinthe.Middleware

  import ArchitectWeb.Helpers.ValidationMessageHelpers
  alias ArchitectWeb.Users.User

  def call(resolution, _config) do
    case resolution.context do
      %{current_user: %User{}} ->
        resolution

      _ ->
        message = "Vous devez vous connecter ou vous inscrire pour continuer."
        resolution |> Absinthe.Resolution.put_result({:ok, generic_message(message)})
    end
  end
end
