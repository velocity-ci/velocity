defmodule Velocity.Repo.Migrations.CreateProjects do
  use Ecto.Migration

  def change do
    create table(:projects) do
      add :id_name, :string
      add :name, :string
      add :repository, :string
      add :key, :string

      timestamps()
    end

  end
end
