defmodule Architect.Projects.Project.NameSlug do
  use EctoAutoslugField.Slug, from: :name, to: :slug
end

defmodule Architect.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Architect.Projects.Repository

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
    |> cast(attrs, [:name, :address, :private_key])
    |> validate_required([:name, :address])
    |> unique_constraint(:name)
    |> NameSlug.maybe_generate_slug()
    |> NameSlug.unique_constraint()
    |> clone()
  end

  @doc """
  Populate a KnownHost changeset by scanning the value specified at host, if changeset is valid
  """
  def clone(%Changeset{valid?: true, changes: %{address: address}} = changeset) do
    require Logger

    repository_name = {:via, Registry, {Architect.Projects.Registry, address}}

    {:ok, repository} =
      DynamicSupervisor.start_child(
        Architect.Projects.Supervisor,
        {Repository, {address, repository_name}}
      )

    if Repository.cloned_successfully(repository) do
      changeset
    else
      add_error(changeset, :address, "Cloning repository failed")
    end
  end

  def clone(changeset), do: changeset
end
