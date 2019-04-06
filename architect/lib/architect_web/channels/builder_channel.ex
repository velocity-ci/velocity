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
  def handle_in("#{@event_prefix}builder-ready", nil, socket) do
    # IO.inspect(Architect.Builds.list_builds())

    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-build events.
  """
  def handle_in("#{@event_prefix}update-build", nil, socket) do
    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-step events.
  """
  def handle_in("#{@event_prefix}update-step", nil, socket) do
    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-stream events.
  """
  def handle_in("#{@event_prefix}update-stream", nil, socket) do
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
        address: "https://github.com/velocity-ci/velocity.git",
        privateKey: ""
      },
      knownHost: %{
        # entry: ""
      },
      # output from vcli plan
      buildTask: %{},
      branch: "master",
      commit: "",
      parameters: %{}
    })

    {:noreply, socket}
  end

  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Builders.track(socket)

    {:noreply, socket}
  end
end
