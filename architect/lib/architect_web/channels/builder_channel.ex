defmodule ArchitectWeb.BuilderChannel do
  use Phoenix.Channel
  alias Architect.Builders

  require Logger

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
  def handle_in("#{@event_prefix}builder-ready", payload, socket) do
    # IO.inspect(Architect.Builds.list_builds())

    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-build events.
  """
  def handle_in("#{@event_prefix}build-stream:new-loglines", payload, socket) do
    # IO.inspect(payload)
    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-step events.
  """
  def handle_in("#{@event_prefix}build-step:update", payload, socket) do
    # IO.inspect(payload)
    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-stream events.
  """
  def handle_in("#{@event_prefix}build-task:update", payload, socket) do
    # IO.inspect(payload)
    {:reply, :ok, socket}
  end

  @doc """
  Starts a build job for a builder.
  """
  def handle_info(b = %Architect.Builds.Build{}, socket) do
    push(socket, "#{@event_prefix}job-do-build", %{
      id: b.id,
      project: %{
        name: b.project.name,
        address: b.project.address,
        privateKey: ""
      },
      knownHost: %{
        # entry: ""
      },
      # output from vcli plan
      buildTask: b.plan,
      branch: b.branch_name,
      commit: b.commit_sha,
      parameters: b.parameters,
    })

    {:noreply, socket}
  end

  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Builders.track(socket)

    {:noreply, socket}
  end
end
