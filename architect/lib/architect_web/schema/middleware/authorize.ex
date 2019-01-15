defmodule ArchitectWeb.Schema.Middleware.Authorize do
  @behaviour Absinthe.Middleware

  def call(resolution, _config) do
    case resolution.context do
      # %{current_user: %User{}} ->
      #   resolution

      _ ->
        resolution

        #        Absinthe.Resolution.put_result(resolution, {:ok, generic_message("not authorized")})
    end
  end
end
