defmodule Architect.Builders.StageScheduler do
  @moduledoc false

  use GenStage
  require Logger
  alias DlqHandler.Message

  @behaviour GenStage

  def child_spec(args) do
    %{
      id: queue,
      start: {__MODULE__, :start_link, args},
      type: :worker
    }
  end

  @doc false
  def start_link(%{task_queue: queue} = queue_opts, _opts) do
    GenStage.start_link(__MODULE__, [queue_opts], name: __MODULE__)
  end

  @doc false
  @impl true
  def init([%{task_queue: queue, repo: repo}]) do
    Process.send(self(), :get_messages, [])

    {:producer, :no_state, demand: :accumulate}
  end

  @impl true
  def handle_demand(new_demand, :no_state) do
    {:noreply, [], :no_state}
  end

  @impl true
  def handle_info(:get_waiting_builds, :no_state) do
    {:ok, builds} = Architect.Builds.list_waiting_builds()

    Process.send_after(self(), :get_waiting_builds, 3_000)

    {:noreply, builds, :no_state, :hibernate}
  end



  defp list_request(queue, wait_time),
       do:
         SQS.receive_message(queue,
           visibility_timeout: 1,
           max_number_of_messages: 10,
           wait_time_seconds: wait_time
         )
         |> ExAws.request()
end
