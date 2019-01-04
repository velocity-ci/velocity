defmodule Architect.KnownHosts do
  @moduledoc """
  The KnownHosts context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.KnownHosts.KnownHost

  @doc """
  Returns the list of known_hosts.

  ## Examples

      iex> list_known_hosts()
      [%KnownHost{}, ...]

  """
  def list_known_hosts do
    Repo.all(KnownHost)
  end

  @doc """
  Gets a single known_host by id.

  Raises `Ecto.NoResultsError` if the Known host does not exist.

  ## Examples

      iex> get_known_host!(123)
      %KnownHost{}

      iex> get_known_host!(456)
      ** (Ecto.NoResultsError)

  """
  def get_known_host!(id), do: Repo.get!(KnownHost, id)

  @doc """
  Gets a single known_host by host.

  Raises `Ecto.NoResultsError` if the Known host does not exist.

  ## Examples

      iex> get_known_host_by_host!("github.com")
      %KnownHost{}

      iex> get_known_host_by_host!("example.com")
      ** (Ecto.NoResultsError)

  """
  def get_known_host_by_host!(host), do: Repo.get_by!(KnownHost, host: host)

  @doc """
  Creates a known_host.

  ## Examples

      iex> create_known_host(%{field: value})
      {:ok, %KnownHost{}}

      iex> create_known_host(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_known_host(attrs \\ %{}) do
    %KnownHost{}
    |> KnownHost.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a known_host.

  ## Examples

      iex> update_known_host(known_host, %{field: new_value})
      {:ok, %KnownHost{}}

      iex> update_known_host(known_host, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_known_host(%KnownHost{} = known_host) do
    known_host
    |> KnownHost.changeset(%{})
    |> Repo.update()
  end

  @doc """
  Deletes a KnownHost.

  ## Examples

      iex> delete_known_host(known_host)
      {:ok, %KnownHost{}}

      iex> delete_known_host(known_host)
      {:error, %Ecto.Changeset{}}

  """
  def delete_known_host(%KnownHost{} = known_host) do
    Repo.delete(known_host)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking known_host changes.

  ## Examples

      iex> change_known_host(known_host)
      %Ecto.Changeset{source: %KnownHost{}}

  """
  def change_known_host(%KnownHost{} = known_host) do
    KnownHost.changeset(known_host, %{})
  end
end
