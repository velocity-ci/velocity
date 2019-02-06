defmodule Architect.Events do
  alias Architect.Repo
  alias Architect.Events.Event
  alias Architect.Projects.Project
  alias Architect.KnownHosts.KnownHost
  alias Architect.Accounts.User
  import Ecto.Query

  @project_events [:project_created]
  @known_host_events [:known_host_created, :known_host_verified]

  def list_events_query() do
    from(event in Event,
      left_join: user in assoc(event, :user),
      left_join: known_host in assoc(event, :known_host),
      left_join: project in assoc(event, :project),
      preload: [user: user, known_host: known_host, project: project],
      order_by: [desc: event.inserted_at]
    )
  end

  @doc """
  Returns the list of events.

  ## Examples

      iex> list_projects()
      [%Project{}, ...]

  """
  def list_events() do
    Repo.all(list_events_query())
  end

  @doc """

  See create_event/1

  """
  def create_event!(user, entity, attrs) do
    case create_event(user, entity, attrs) do
      {:ok, event} ->
        event

      {:error, error} ->
        throw(error)
    end
  end

  @doc """
  Create an event for a user

  ## Examples

      iex> create_event(user, project, %{type: :project_created})
      {:ok, %Event{}}

  """
  def create_event(%User{} = u, entity, attrs) do
    attrs
    |> Map.put(:user_id, u.id)
    |> create_event(entity)
  end

  def create_event(%{type: type} = attrs, %Project{id: project_id})
      when type in @project_events do
    attrs
    |> Map.put(:project_id, project_id)
    |> create_event()
  end

  def create_event(%{type: type} = attrs, %KnownHost{id: known_host_id})
      when type in @known_host_events do
    attrs
    |> Map.put(:known_host_id, known_host_id)
    |> create_event()
  end

  def create_event(attrs) do
    %Event{}
    |> Event.changeset(attrs)
    |> Repo.insert()
  end
end
