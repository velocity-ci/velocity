defmodule ArchitectWeb.V1.BuilderChannel do
  use Phoenix.Channel
  alias Architect.Builders
  alias Phoenix.Socket

  def join("builders:pool", _, socket) do
    send(self(), :after_join)

    {:ok, socket}
  end

  #  def set_ready(%Socket) when is_pid(pid) do
  #    GenServer.
  #  end

  #  @doc ~S"""
  #  Used to syncronize the client with the orderbook state.
  #
  #  Get all updates from sequence number given to current state
  #  """
  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Builders.track(socket)

    {:noreply, socket}
  end
end
