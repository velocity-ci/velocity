defmodule Architect.Repo.Migrations.CreateProjects do
  use Ecto.Migration

  def change do
    create table(:projects, primary_key: false) do
      add(:id, :uuid, primary_key: true)
      add(:name, :string)
      add(:repository, :string)
      add(:slug, :string)

      timestamps()
    end

    create(unique_index(:projects, [:name]))
    create(unique_index(:projects, [:slug]))
  end
end
