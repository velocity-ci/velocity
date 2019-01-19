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

  def handle_info(:get_commits, socket) do
    push(socket, "vlcty_repo-get-commits", %{
      repository: %{
        address: "https://github.com/velocity-ci/velocity.git",
        knownHostEntry:
          "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
      }
    })

    {:noreply, socket}
  end
end
