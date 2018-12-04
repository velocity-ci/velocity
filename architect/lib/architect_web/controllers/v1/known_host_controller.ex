defmodule ArchitectWeb.V1.KnownHostController do
  use ArchitectWeb, :controller

  alias Architect.KnownHosts
  alias Architect.KnownHosts.KnownHost

  action_fallback ArchitectWeb.V1.FallbackController

  def index(conn, _params) do
    known_hosts = KnownHosts.list_known_hosts()
    render(conn, "index.json", known_hosts: known_hosts)
  end

  def create(conn, %{"known_host" => known_host_params}) do
    with {:ok, %KnownHost{} = known_host} <- KnownHosts.create_known_host(known_host_params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", V1Routes.known_host_path(conn, :show, known_host))
      |> render("show.json", known_host: known_host)
    end
  end

  def show(conn, %{"id" => id}) do
    known_host = KnownHosts.get_known_host!(id)
    render(conn, "show.json", known_host: known_host)
  end

  def update(conn, %{"id" => id, "known_host" => known_host_params}) do
    known_host = KnownHosts.get_known_host!(id)

    with {:ok, %KnownHost{} = known_host} <-
           KnownHosts.update_known_host(known_host, known_host_params) do
      render(conn, "show.json", known_host: known_host)
    end
  end

  def delete(conn, %{"id" => id}) do
    known_host = KnownHosts.get_known_host!(id)

    with {:ok, %KnownHost{}} <- KnownHosts.delete_known_host(known_host) do
      send_resp(conn, :no_content, "")
    end
  end
end
