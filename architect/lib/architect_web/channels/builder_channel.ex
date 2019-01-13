defmodule ArchitectWeb.BuilderChannel do
  use Phoenix.Channel
  alias Architect.Builders

  @event_prefix "vlcty_"

  def join("builders:pool", _, socket) do
    send(self(), :after_join)

    {:ok, socket}
  end

  def handle_in("new_msg", %{"uid" => _uid, "body" => _body}, socket) do
    # broadcast!(socket, "new_msg", %{uid: uid, body: body})

    push(socket, "", %{})

    {:noreply, socket}
  end

  @doc """
  Builder will say when it is 'ready' to request any waiting jobs.
  """
  def handle_in("#{@event_prefix}builder-ready", nil, socket) do
    {:noreply, socket}
  end

  @doc """
  Starts a synchronisation job for a builder.
  """
  def handle_info(:job_synchronise, socket) do
    push(socket, "#{@event_prefix}do-synchronise", %{})

    {:noreply, socket}
  end

  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Builders.track(socket)

    {:noreply, socket}
  end
end
