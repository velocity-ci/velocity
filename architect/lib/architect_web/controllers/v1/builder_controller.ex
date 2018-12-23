defmodule ArchitectWeb.V1.BuilderController do
  use ArchitectWeb, :controller

  action_fallback(ArchitectWeb.V1.FallbackController)

  def create(conn, %{"secret" => secret}) do
    if secret == Application.get_env(:architect, :builder_secret) do
      with {:ok, builder} <- Architect.Builders.create_builder() do
        conn
        |> put_status(:created)
        |> put_resp_header("location", V1Routes.builder_path(conn, :show, builder))
        |> render("show.json", builder: builder)
      end
    else
      {:error, :invalid_credentials}
    end
  end
end
