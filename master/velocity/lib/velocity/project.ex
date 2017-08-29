defmodule Velocity.Project do
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Velocity.Project
  alias Velocity.Repo

  schema "projects" do
    field :id_name, :string
    field :key, :string
    field :name, :string
    field :repository, :string

    timestamps()
  end

  @doc false
  def changeset(%Project{} = project, attrs) do
    project
    |> cast(attrs, [:id_name, :name, :repository, :key])
  end

  @spec create(Ecto.Changeset) :: struct
  def create(changeset) do
    changeset
    |> put_change(:id_name, String.downcase(String.replace(changeset.params["name"], " ", "-")))
    |> Repo.insert
  end
end
