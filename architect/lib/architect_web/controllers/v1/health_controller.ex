defmodule ArchitectWeb.V1.HealthController do
  use ArchitectWeb, :controller

  def index(conn, _params) do
    render(conn, "index.json")
  end
end
