defmodule Architect.KnownHosts do
  @moduledoc """
  The KnownHosts context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.KnownHosts.KnownHost

  @doc """
  Returns the list of known_hosts.
  """
  def list_known_hosts do
    Repo.all(KnownHost)
  end

  @doc """
  Gets a single known_host by id.

  Raises `Ecto.NoResultsError` if the Known host does not exist.
  """
  def get_known_host!(id), do: Repo.get!(KnownHost, id)

  @doc """
  Gets a single known_host by host.

  Raises `Ecto.NoResultsError` if the Known host does not exist.
  """
  def get_known_host_by_host!(host), do: Repo.get_by!(KnownHost, host: host)

  @doc """
  Creates a known_host.

  """
  def create_known_host(attrs \\ %{}) do
    %KnownHost{}
    |> KnownHost.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a known_host.

  """
  def update_known_host(%KnownHost{} = known_host, attrs) do
    known_host
    |> KnownHost.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a KnownHost.
  """
  def delete_known_host(%KnownHost{} = known_host) do
    Repo.delete(known_host)
  end
end
