defmodule ArchitectWeb.V1.SecretaryChannel do
  use Phoenix.Channel
  alias Architect.Secretaries

  @event_prefix "vlcty_"

  def join("secretaries:pool", _, socket) do
    send(self(), :after_join)

    {:ok, socket}
  end

  @spec handle_info(:after_join | :job_synchronise, Phoenix.Socket.t()) :: {:noreply, map()}
  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Secretaries.track(socket)

    {:noreply, socket}
  end

  def handle_in("new_msg", %{"uid" => uid, "body" => body}, socket) do
    # broadcast!(socket, "new_msg", %{uid: uid, body: body})

    push(socket, "", %{})

    {:noreply, socket}
  end

  def handle_info(:health_check, socket) do
    push(socket, "vlcty_health-check", %{})
    {:noreply, socket}
  end
end
