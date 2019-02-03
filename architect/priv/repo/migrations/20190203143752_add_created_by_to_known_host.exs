defmodule Architect.Repo.Migrations.AddCreatedByToKnownHost do
  use Ecto.Migration

  def change do

    alter table(:known_hosts) do
      add(
        :created_by_id,
        references(:users, type: :uuid)
      )
    end

  end
end
