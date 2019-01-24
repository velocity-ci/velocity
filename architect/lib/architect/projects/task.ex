defmodule Architect.Projects.Task do


  defstruct [:name, :description, :git, :docker, :parameters]
  @enforce_keys [:name, :description]

  @doc ~S"""

  ## Example

    iex> Architect.Projects.Task.parse(%{"name" => "build-dev", "description" => "This builds dev"})
    %Architect.Projects.Task{name: "build-dev", description: "This builds dev"}

  """
  def parse(%{"name" => name, "description" => description}) do
    %__MODULE__{name: name, description: description}
  end

end