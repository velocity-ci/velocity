defmodule ArchitectWeb.V1.BuilderChannel do
  use Phoenix.Channel

  def join("builder:" <> builder_id, params, socket) do
    # TODO: authenticate builder token with registry
    {:ok, socket}
    # {:error, %{reason: "unauthorized"}}
  end
end
