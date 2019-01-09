defmodule ArchitectWeb.Helpers.ValidationMessageHelpers do
  alias Kronky.ValidationMessage

  def generic_message(message) when is_binary(message) do
    %ValidationMessage{
      code: :unknown,
      field: "base",
      key: 0,
      template: message,
      message: message,
      options: []
    }
  end

  def message(field, message) when is_binary(message) do
    %ValidationMessage{
      code: :unknown,
      field: field,
      key: 0,
      template: message,
      message: message,
      options: []
    }
  end
end
