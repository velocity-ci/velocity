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
  def handle_in("#{@event_prefix}task-stream:new-loglines", payload, socket) do
    Enum.each(payload["lines"], fn l ->
      Architect.Builds.ETSStore.put_stream_line(
        payload["id"],
        l["lineNumber"],
        l
      )
    end)

    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-step events.
  """
  def handle_in("#{@event_prefix}task-step:update", payload, socket) do
    Architect.Builds.ETSStore.put_step_update(
      payload["id"],
      payload
    )

    {:reply, :ok, socket}
  end

  @doc """
  Handle build update-stream events.
  """
  def handle_in("#{@event_prefix}task-task:update", payload, socket) do
    Architect.Builds.ETSStore.put_task_update(
      payload["id"],
      payload
    )

    {:reply, :ok, socket}
  end

  @doc """
  Starts a task on a builder.
  """
  def handle_info({b = %Architect.Builds.Build{}, t = %Architect.Builds.Task{}}, socket) do
    push(socket, "#{@event_prefix}job-do-task", %{
      id: b.id,
      project: %{
        name: b.project.name,
        address: b.project.address,
        privateKey: ""
      },
      knownHost:
        %{
          # entry: ""
        },
      task: t.plan,
      branch: b.branch_name,
      commit: b.commit_sha,
      parameters: b.parameters
    })

    {:noreply, socket}
  end

  def handle_info(:after_join, socket) do
    socket = assign(socket, :status, :ready)

    {:ok, _} = Builders.track(socket)

    {:noreply, socket}
  end
end
