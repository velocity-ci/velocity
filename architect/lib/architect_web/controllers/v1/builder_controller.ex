defmodule ArchitectWeb.V1.BuilderController do
  use ArchitectWeb, :controller

  def create(conn, %{"known_host" => known_host_params}) do
    # with {:ok, %KnownHost{} = known_host} <- KnownHosts.create_known_host(known_host_params) do
    #   conn
    #   |> put_status(:created)
    #   |> put_resp_header("location", V1Routes.known_host_path(conn, :show, known_host))
    #   |> render("show.json", known_host: known_host)
    # end
  end
end
