defmodule Architect.Projects.Project.NameSlug do
  use EctoAutoslugField.Slug, from: :name, to: :slug
end

defmodule Architect.Projects.Project do
  use Ecto.Schema
  import Ecto.Changeset

  alias Comeonin.Bcrypt
  alias __MODULE__.NameSlug

  @primary_key {:id, :binary_id, autogenerate: true}
  schema "projects" do
    field(:name, :string)
    field(:repository, :string)
    field(:slug, NameSlug.Type)

    timestamps()
  end

  @doc false
  def changeset(project, attrs) do
    project
    |> cast(attrs, [:name, :repository])
    |> validate_required([:name, :repository])
    |> unique_constraint(:name)
    |> NameSlug.maybe_generate_slug()
    |> NameSlug.unique_constraint()
  end
end
