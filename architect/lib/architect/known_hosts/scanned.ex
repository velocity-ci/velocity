defmodule Architect.KnownHosts.Scanned do
  @moduledoc """
  Provides functions to get known host data from a host.

  Does this by running the v-ssh-keyscan executable and decoding the output to json.

  Tries to scan for *:architect, :keyscan_timeout* milliseconds, then will fail with :keyscan_timeout.

  Possible errors:

    :invalid_args - A non-binary was passed
    :keyscan_timeout - No response after 5 seconds from keyscan executable
    :keyscan_failed - The keyscan executable returned a non 0 exit code
    :json_decode_failed - Could not decode the executable output to JSON
    :unexpected_decode_values - Could decode the executable output to JSON, but could not find required keys

  """
  require Logger

  @enforce_keys [:md5, :sha256, :entry]
  defstruct [:md5, :sha256, :entry]

  @keyscan_bin "v-ssh-keyscan"

  @doc """
  Scan a host and get either a :ok or :error tuple

  Examples

      ...> Architect.KnownHosts.Scanned.generate("github.com")
      {:ok, Architect.KnownHosts.Scanned{}}

  """
  def generate(host) when is_binary(host) do
    %{timeout: timeout} =
      Application.get_env(:architect, :keyscan)
      |> Enum.into(%{})

    try do
      Task.async(fn ->
        System.cmd("#{System.cwd()}/#{@keyscan_bin}", [host], stderr_to_stdout: true)
      end)
      |> Task.await(timeout)
      |> handle_scan()
    catch
      :exit, _ ->
        log("v-ssh-keyscan timeout", :warn)
        {:error, :keyscan_timeout}
    end
  end

  def generate(_), do: {:error, :invalid_args}

  @doc false
  defp handle_scan({output, exit_code}) when exit_code == 0 do
    output
    |> Poison.decode()
    |> handle_decode()
  end

  defp handle_scan({output, exit_code}) do
    log("v-ssh-keyscan exit code #{inspect(exit_code)}, error: #{inspect(output)}", :error)

    {:error, :keyscan_failed}
  end

  @doc false
  defp handle_decode(
         {:ok,
          %{
            "entry" => entry,
            "sha256Fingerprint" => sha256,
            "md5Fingerprint" => md5
          }}
       ) do
    {:ok, %__MODULE__{md5: md5, sha256: sha256, entry: entry}}
  end

  defp handle_decode({:ok, values}) do
    log("v-ssh-keyscan unexpected decode values #{inspect(values)}", :error)

    {:error, :unexpected_decode_values}
  end

  defp handle_decode({:error, values}) do
    log("v-ssh-keyscan failed to decode to JSON #{inspect(values)}", :error)

    {:error, :json_decode_failed}
  end

  @doc false
  defp log(output, level) do
    %{log_errors: log_errors} =
      Application.get_env(:architect, :keyscan)
      |> Enum.into(%{})

    log(output, level, log_errors)
  end

  defp log(output, :debug, true), do: Logger.debug(output)
  defp log(output, :warn, true), do: Logger.warn(output)
  defp log(output, :error, true), do: Logger.error(output)
  defp log(_, _, _), do: nil
end
