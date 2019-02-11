defmodule Architect.KnownHosts do
  @moduledoc """
  The KnownHosts context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.KnownHosts.KnownHost
  alias Architect.Accounts.User
  alias Architect.Events

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
  def create_known_host(%User{} = u, host) when is_binary(host) do
    Repo.transaction(fn ->
      changeset = KnownHost.changeset(%KnownHost{}, %{host: host, created_by_id: u.id})

      case Repo.insert(changeset) do
        {:ok, known_host} ->
          event = Events.create_event!(u, known_host, %{type: :known_host_created})

          {known_host, event}

        {:error, e} ->
          Repo.rollback(e)
      end
    end)
  end

  @doc """
  Verify a known_host.

  """
  def verify_known_host(%User{} = u, %KnownHost{} = known_host) do
    Repo.transaction(fn ->
      changeset = KnownHost.changeset(known_host, %{verified: true})

      case Repo.update(changeset) do
        {:ok, k} ->
          event = Events.create_event!(u, k, %{type: :known_host_verified})

          {k, event}

        {:error, e} ->
          Repo.rollback(e)
      end
    end)
  end

  @doc """
  Deletes a KnownHost.
  """
  def delete_known_host(%KnownHost{} = known_host) do
    Repo.delete(known_host)
  end
end
