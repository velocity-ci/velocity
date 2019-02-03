defmodule Architect.Repo.Migrations.ProjectCreatedByUser do
  use Ecto.Migration

  def change do

    alter table(:projects) do
      add(
        :created_by_id,
        references(:users, type: :uuid)
      )
    end

  end
end
