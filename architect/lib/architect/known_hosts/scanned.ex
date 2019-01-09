defmodule Architect.KnownHosts.Scanned do
  @moduledoc """
  Provides functions to get known host data from a host.

  Does this by running the v-ssh-keyscan executable and decoding the output to json.

  Tries to scan for 20 seconds, then will fail with :keyscan_timeout.

  Possible errors:

    :invalid_args - A non-binary was passed
    :keyscan_timeout - No response after 5 seconds from keyscan executable
    :keyscan_failed - The keyscan executable returned a non 0 exit code
    :json_decode_failed - Could not decode the executable output to JSON
    :unexpected_decode_values - Could decode the executable output to JSON, but could not find required keys

  ## Examples

      iex> Architect.KnownHosts.Scanned.generate("github.com")
      {:ok,
        %Architect.KnownHosts.Scanned{
         entry: "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==\n",
         md5: "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48",
         sha256: "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"
        }
      }

  """
  require Logger

  @enforce_keys [:md5, :sha256, :entry]
  defstruct [:md5, :sha256, :entry]

  @keyscan_bin "v-ssh-keyscan"
  @timeout 7_000

  @doc """
  Scan a host and get either a :ok or :error tuple
  """
  def generate(host) when is_binary(host) do
    try do
      Task.async(fn ->
        System.cmd("#{System.cwd()}/#{@keyscan_bin}", [host], stderr_to_stdout: true)
      end)
      |> Task.await(@timeout)
      |> handle_scan()
    catch
      :exit, _ ->
        Logger.error("v-ssh-keyscan timeout")
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
    Logger.error("v-ssh-keyscan exit code #{inspect(exit_code)}, error: #{inspect(output)}")

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
    Logger.error("v-ssh-keyscan unexpected decode values #{inspect(values)}")

    {:error, :unexpected_decode_values}
  end

  defp handle_decode({:error, values}) do
    Logger.error("v-ssh-keyscan failed to decode to JSON #{inspect(values)}")

    {:error, :json_decode_failed}
  end
end
