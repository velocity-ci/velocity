defmodule Architect.Projects.Project.NameSlug do
  use EctoAutoslugField.Slug, from: :name, to: :slug
end

defmodule Architect.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Git.Repository

  alias __MODULE__.NameSlug

  @primary_key {:id, :binary_id, autogenerate: true}
  schema "projects" do
    field(:name, :string)
    field(:address, :string)
    field(:private_key, :string)

    field(:slug, NameSlug.Type)

    timestamps()
  end

  @doc false
  def changeset(project, attrs) do
    project
    |> cast(attrs, [:address, :private_key])
    |> validate_required([:address])
    |> update_default_name()
    |> unique_constraint(:address)
    |> NameSlug.maybe_generate_slug()
    |> NameSlug.unique_constraint()
    |> verify()
  end

  @doc ~S"""
  If address has changed, update the name of the project
  """
  def update_default_name(%Changeset{changes: %{address: address}} = changeset) do
    put_change(changeset, :name, default_name(address))
  end

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

  @doc """
  Start the project repository, if it fails we add an error to the changeset and terminate the process.

  This means on a successful verify, the repository process is already ready to go
  """
  def verify(
        %Changeset{
          valid?: true,
          changes: %{address: address, name: name}
        } = changeset
      ) do
    require Logger

    private_key = Changeset.get_change(changeset, :private_key)

    repository_name = {:via, Registry, {Architect.Projects.Registry, "#{address}-#{name}"}}
    IO.inspect(repository_name)
    IO.inspect(address)
    IO.inspect(name)

    repository_result =
      DynamicSupervisor.start_child(
        Architect.Projects.Supervisor,
        {Repository, {address, private_key, repository_name}}
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
      :ok = DynamicSupervisor.terminate_child(Architect.Projects.Supervisor, repository)
      add_error(changeset, :address, "Verifying repository failed")
    end
  end
end
