defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository
  """

  defstruct [:repo]
  use GenServer
  require Logger

  # Client

  def start_link(url) when is_binary(url) do
    Logger.debug("Starting fresh repository process for #{url}")

    GenServer.start_link(__MODULE__, url)
  end

  # Server (callbacks)

  @impl true
  def init(url) when is_binary(url) do
    uuid = UUID.uuid4()
    path = temp_path(uuid)
    with {:ok, repo} <- Git.clone([url, uuid]) do
      Logger.debug("Successfully cloned #{url} to #{path}")
      {:ok, %__MODULE__{repo: repo}}
    else
      {:error, %Git.Error{message: reason}} ->
        {:stop, reason}
    end
  end



  defp temp_path(uuid) when is_binary(uuid) do
    Temp.track!
    Temp.mkdir!(uuid)
  end

end