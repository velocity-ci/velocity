defmodule VelocityWeb.PageController do
  @moduledoc "provides the index"
  use VelocityWeb, :controller

  def index(conn, _params) do
    render conn, "index.html"
  end
end
