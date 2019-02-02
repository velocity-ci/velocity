defmodule Architect.Projects.Task do
  @moduledoc """
  Represents a task in Velocity

  Currently only root level properties are parsed.

  The properties in this root struct contain raw maps with string keys for all properties
  """

  defstruct [:name, :description, :git, :docker, :parameters]
  @enforce_keys [:name, :description]

  @doc ~S"""

  ## Example

    iex> Architect.Projects.Task.parse(%{"name" => "build-dev", "description" => "This builds dev"})
    %Architect.Projects.Task{name: "build-dev", description: "This builds dev"}

  """
  def parse(output) when is_list(output) do
    Enum.map(output, &parse/1)
  end

  def parse(%{"name" => name, "description" => description} = output) do
    %__MODULE__{name: name, description: description}
    |> parse_git(output)
    |> parse_docker(output)
    |> parse_parameters(output)
  end

  @doc ~S"""
  Adds git to task

  ## Example

    iex> task = %Architect.Projects.Task{name: "test", description: "test-desc"}
    ...> Architect.Projects.Task.parse_git(task, %{"git" => %{submodule: false}})
    %Architect.Projects.Task{name: "test", description: "test-desc", git: %{submodule: false}}

  """
  def parse_git(%__MODULE__{} = task, %{"git" => git}), do: %{task | git: git}

  def parse_git(task, _), do: task

  @doc ~S"""
  Adds docker to task

  ## Example

    iex> task = %Architect.Projects.Task{name: "test", description: "test-desc"}
    ...> Architect.Projects.Task.parse_docker(task, %{"docker" => %{registries: []}})
    %Architect.Projects.Task{name: "test", description: "test-desc", docker: %{registries: []}}

  """
  def parse_docker(%__MODULE__{} = task, %{"docker" => docker}), do: %{task | docker: docker}

  def parse_docker(task, _), do: task

  @doc ~S"""
  Adds parameters to task

  ## Example

    iex> task = %Architect.Projects.Task{name: "test", description: "test-desc"}
    ...> Architect.Projects.Task.parse_docker(task, %{"parameters" => %{parameters: []}})
    %Architect.Projects.Task{name: "test", description: "test-desc", docker: %{parameters: []}}

  """
  def parse_parameters(%__MODULE__{} = task, %{"parameters" => parameters}),
    do: %{task | parameters: parameters}

  def parse_parameters(task, _), do: task
end
