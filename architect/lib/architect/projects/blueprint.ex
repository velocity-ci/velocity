defmodule Architect.Projects.Blueprint do
  @moduledoc """
  Represents a blueprint in Velocity

  Currently only root level properties are parsed.

  The properties in this root struct contain raw maps with string keys for all properties
  """

  defstruct [:name, :description, :git, :docker, :parameters]
  @enforce_keys [:name, :description]

  @doc ~S"""

  ## Example

    iex> Architect.Projects.Blueprint.parse(%{"name" => "build-dev", "description" => "This builds dev"})
    %Architect.Projects.Blueprint{name: "build-dev", description: "This builds dev"}

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
  Adds git to blueprint

  ## Example

    iex> blueprint = %Architect.Projects.Blueprint{name: "test", description: "test-desc"}
    ...> Architect.Projects.Blueprint.parse_git(blueprint, %{"git" => %{submodule: false}})
    %Architect.Projects.Blueprint{name: "test", description: "test-desc", git: %{submodule: false}}

  """
  def parse_git(%__MODULE__{} = blueprint, %{"git" => git}), do: %{blueprint | git: git}

  def parse_git(blueprint, _), do: blueprint

  @doc ~S"""
  Adds docker to blueprint

  ## Example

    iex> blueprint = %Architect.Projects.Blueprint{name: "test", description: "test-desc"}
    ...> Architect.Projects.Blueprint.parse_docker(blueprint, %{"docker" => %{registries: []}})
    %Architect.Projects.Blueprint{name: "test", description: "test-desc", docker: %{registries: []}}

  """
  def parse_docker(%__MODULE__{} = blueprint, %{"docker" => docker}), do: %{blueprint | docker: docker}

  def parse_docker(blueprint, _), do: blueprint

  @doc ~S"""
  Adds parameters to blueprint

  ## Example

    iex> blueprint = %Architect.Projects.Blueprint{name: "test", description: "test-desc"}
    ...> Architect.Projects.Blueprint.parse_docker(blueprint, %{"parameters" => %{parameters: []}})
    %Architect.Projects.Blueprint{name: "test", description: "test-desc", docker: %{parameters: []}}

  """
  def parse_parameters(%__MODULE__{} = blueprint, %{"parameters" => parameters}),
    do: %{blueprint | parameters: parameters}

  def parse_parameters(blueprint, _), do: blueprint
end
