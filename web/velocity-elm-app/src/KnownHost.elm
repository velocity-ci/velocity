module KnownHost exposing (KnownHost, addKnownHost, decoder, list)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import GitUrl exposing (GitUrl)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, hardcoded, required)


type alias KnownHost =
    { id : Id
    , hosts : List String
    , comment : String
    , sha256 : String
    , md5 : String
    }



-- SERIALIZATION --


decoder : Decoder KnownHost
decoder =
    Decode.succeed KnownHost
        |> required "id" decodeId
        |> required "hosts" (Decode.list Decode.string)
        |> required "comment" Decode.string
        |> required "sha256" Decode.string
        |> required "md5" Decode.string



-- HELPERS --


isUnknownHost : List KnownHost -> Maybe GitUrl -> Bool
isUnknownHost knownHosts maybeGitUrl =
    case maybeGitUrl of
        Just { source } ->
            knownHosts
                |> hostsFromKnownHosts
                |> List.member source
                |> not

        Nothing ->
            False


hostsFromKnownHosts : List KnownHost -> List String
hostsFromKnownHosts knownHosts =
    List.foldl (.hosts >> (++)) [] knownHosts


findKnownHost : List KnownHost -> KnownHost -> Maybe KnownHost
findKnownHost knownHosts knownHost =
    List.filter (\a -> a.id == knownHost.id) knownHosts
        |> List.head


addKnownHost : List KnownHost -> KnownHost -> List KnownHost
addKnownHost knownHosts knownHost =
    case findKnownHost knownHosts knownHost of
        Just _ ->
            knownHosts

        Nothing ->
            knownHost :: knownHosts



-- IDENTIFIERS --


type Id
    = Id String


decodeId : Decoder Id
decodeId =
    Decode.map Id Decode.string


idToString : Id -> String
idToString (Id id) =
    id



-- COLLECTION --


list : Maybe Cred -> BaseUrl -> Http.Request (List KnownHost)
list maybeCred baseUrl =
    let
        endpoint =
            Endpoint.knownHosts { amount = -1, page = 1 } (Api.toEndpoint baseUrl)
    in
    Decode.field "data" (Decode.list decoder)
        |> Api.get endpoint maybeCred
