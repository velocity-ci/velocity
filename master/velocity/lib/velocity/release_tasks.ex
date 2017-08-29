defmodule Velocity.ReleaseTasks do
  @moduledoc "provides init tasks for releases."

  alias Ecto.Migrator

  @start_apps [:postgrex, :ecto]
  @repos [Velocity.Repo]
  def migrations do
    # Load the code for Velocity, but don't start it
    IO.write "> Loading Velocity..."
    :ok = Application.load(:velocity)
    IO.puts " Done."
    # Start apps necessary for executing migrations
    IO.write "> Starting dependencies..."
    Enum.each @start_apps, &Application.ensure_all_started/1
    IO.puts " Done."
    # Start the Repo(s) for velocity
    IO.write "> Starting repositories..."
    Enum.each @repos, &apply(&1, :start_link, [])
    IO.puts " Done."
    # Run migrations
    Enum.each @repos, &run_migrations_for/1
    :init.stop
  end

  def create_admin do
    # Create admin user if neccessary
    unless Velocity.UserRepository.find_by_username("admin") do
      Velocity.Repo.insert!(%Velocity.User{username: "admin", password: "password"})
      IO.puts "Created username: admin, password: password"
    end
  end

  def priv_dir do
    Application.app_dir :velocity, "priv"
  end

  defp run_migrations_for(app) do
    IO.puts "> Running migrations for #{app}..."
    Migrator.run app, migrations_path(), :up, all: true
  end

  defp migrations_path do
    Path.join [priv_dir(), "repo", "migrations"]
  end
end
