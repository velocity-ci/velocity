defmodule Architect.Builders.Presence do
  use Phoenix.Presence,
    otp_app: :architect,
    pubsub_server: Architect.PubSub

  alias Phoenix.Socket

  alias __MODULE__

  def topic(), do: "builder_presence"

  @doc """
  Track a user on the @presence_topic by their username
  """
  def track(%Socket{assigns: %{id: id, status: status}, channel_pid: pid}) do
    Presence.track(pid, topic(), id, %{
      online_at: inspect(System.system_time(:second)),
      status: status,
      socket: self()
    })
  end

  def update_status(%Socket{assigns: %{id: id}, channel_pid: pid}, status) do
    Presence.update(pid, topic(), id, fn x -> Map.put(x, :status, status) end)
  end

  def list(), do: Presence.list(topic())
end
