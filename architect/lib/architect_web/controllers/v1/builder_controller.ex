defmodule ArchitectWeb.V1.BuilderController do
  use ArchitectWeb, :controller

  def create(conn, %{"secret" => _secret}) do
    # TODO: verify secret
    with {:ok, builder} <- Architect.Builders.create_builder() do
      conn
      |> put_status(:created)
      |> put_resp_header("location", V1Routes.builder_path(conn, :show, builder))
      |> render("show.json", builder: builder)
    end
  end
end
