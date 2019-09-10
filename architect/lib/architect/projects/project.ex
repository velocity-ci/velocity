defmodule Architect.Projects.Project.NameSlug do
  use EctoAutoslugField.Slug, from: :name, to: :slug
end

defmodule Architect.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Git.Repository
  alias Architect.Accounts.User
  alias Architect.KnownHosts

  alias __MODULE__.NameSlug

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "projects" do
    field(:name, :string)
    field(:address, :string)
    field(:private_key, :string)

    field(:slug, NameSlug.Type)

    belongs_to(:created_by, User)
    timestamps()
  end

  @doc false
  def changeset(%__MODULE__{} = project, attrs) do
    project
    |> cast(attrs, [:address, :private_key, :created_by_id])
    |> assoc_constraint(:created_by)
    |> validate_required([:address])
    |> unique_constraint(:address)
    |> update_default_name()
    |> unique_constraint(:name)
    |> NameSlug.maybe_generate_slug()
    |> NameSlug.unique_constraint()
    |> verify()
  end

  defp update_default_name(%Changeset{changes: %{address: address}} = changeset) do
    put_change(changeset, :name, default_name(address))
  end

  defp update_default_name(changeset), do: changeset

  @doc ~S"""
  Generate a name for the project based on its repository address

  ## Examples

      iex> Architect.Projects.Project.default_name("http://github.com/foo/bar.git")
      "foo/bar @ github.com"

      iex> Architect.Projects.Project.default_name("https://github.com/foo/bar.git")
      "foo/bar @ github.com"

      iex> Architect.Projects.Project.default_name("git@github.com:foo/bar.git")
      "foo/bar @ github.com"

  """
  def default_name("http" <> _ = address) do
    [_proto, host, path] = String.split(address, "/", parts: 3, trim: true)
    path = String.trim_trailing(path, ".git")
    "#{path} @ #{host}"
  end

  def default_name("git" <> _ = address) do
    [_, name] = String.split(address, "@")
    [host, path] = String.split(name, ":")
    path = String.trim_trailing(path, ".git")
    "#{path} @ #{host}"
  end

  def known_hosts("http" <> _ = address) do
    nil
  end

  def known_hosts("git" <> _ = address) do
    [_, name] = String.split(address, "@")
    [host, _path] = String.split(name, ":")
    kh = KnownHosts.get_known_host_by_host!(host)

    case kh.verified do
      true -> kh.entry
      _ -> nil
    end
  end

  @doc """
  Start the project repository, if it fails we add an error to the changeset and terminate the process.

  This means on a successful verify, the repository process is already ready to go
  """
  def verify(
        %Changeset{
          valid?: true,
          changes: %{address: address, created_by_id: created_by_id, slug: slug}
        } = changeset
      ) do
    require Logger

    private_key = Changeset.get_field(changeset, :private_key)
    known_hosts = known_hosts(address)

    process_name = {:via, Registry, {Architect.Projects.Registry, slug}}

    repository_result =
      DynamicSupervisor.start_child(
        Architect.Projects.Supervisor,
        {Repository, {process_name, address, private_key, known_hosts}}
      )

    case repository_result do
      {:ok, repository} ->
        verify(changeset, repository)

      {:error, {:already_started, repository}} ->
        verify(changeset, repository)
    end
  end

  def verify(changeset), do: changeset

  def verify(%Changeset{} = changeset, repository) when is_pid(repository) do
    if Repository.verified?(repository) do
      changeset
    else
      DynamicSupervisor.terminate_child(Architect.Projects.Supervisor, repository)
      add_error(changeset, :address, "Verifying repository failed")
    end
  end
end
