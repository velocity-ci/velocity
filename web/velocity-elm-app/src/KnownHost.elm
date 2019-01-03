module KnownHost exposing (KnownHost, CreateResponse(..), addKnownHost, create, findKnownHost, isUnknownHost, list, selectionSet, createUnverified)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import GitUrl exposing (GitUrl)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline as Pipeline exposing (custom, hardcoded, required)
import Json.Encode as Encode
import Task exposing (Task)
import Graphql.Http
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet, hardcoded, with)
import Graphql.Operation exposing (RootQuery)
import Api.Compiled.Object.KnownHost as KnownHost
import Api.Compiled.Object
import Api.Compiled.Scalar as Scalar
import Api.Compiled.Query as Query
import Api.Compiled.Object.KnownHostPayload as KnownHostPayload
import Api.Compiled.Mutation as Mutation


type KnownHost
    = KnownHost Internals


type alias Internals =
    { id : Id
    , host : String
    , md5 : String
    , sha256 : String
    }



-- SERIALIZATION --


decoder : Decoder KnownHost
decoder =
    Decode.succeed KnownHost
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" decodeId
        |> required "hosts" Decode.string
        |> required "md5" Decode.string
        |> required "sha256" Decode.string


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.KnownHost
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with idSelectionSet
        |> with KnownHost.host
        |> with KnownHost.fingerprintMd5
        |> with KnownHost.fingerprintSha256


selectionSet : SelectionSet KnownHost Api.Compiled.Object.KnownHost
selectionSet =
    SelectionSet.succeed KnownHost
        |> with internalSelectionSet



--
-- INFO


hosts : KnownHost -> List String
hosts (KnownHost knownHost) =
    [ knownHost.host ]



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
    List.foldl (hosts >> (++)) [] knownHosts


findKnownHost : List KnownHost -> KnownHost -> Maybe KnownHost
findKnownHost knownHosts (KnownHost knownHost) =
    List.filter (\(KnownHost a) -> a.id == knownHost.id) knownHosts
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


idSelectionSet : SelectionSet Id Api.Compiled.Object.KnownHost
idSelectionSet =
    SelectionSet.map (\(Scalar.Id id) -> Id id) KnownHost.id



-- REQUESTS


type CreateResponse
    = CreateSuccess KnownHost
    | ValidationFailure (List Api.ValidationMessage)
    | UnknownError


createResponseSelectionSet : SelectionSet CreateResponse Api.Compiled.Object.KnownHostPayload
createResponseSelectionSet =
    let
        messageSelectionSet =
            KnownHostPayload.messages Api.validationErrorSelectionSet
                |> SelectionSet.withDefault []
                |> SelectionSet.nonNullElementsOrFail

        toResponse messages result =
            case result of
                Just knownHost ->
                    CreateSuccess knownHost

                Nothing ->
                    ValidationFailure messages
    in
        SelectionSet.succeed toResponse
            |> SelectionSet.with messageSelectionSet
            |> SelectionSet.with (KnownHostPayload.result selectionSet)


createUnverified : Cred -> BaseUrl -> Graphql.Http.Request (Maybe CreateResponse)
createUnverified cred baseUrl =
    let
        endpoint =
            Api.toEndpoint baseUrl
                |> Endpoint.unwrap
    in
        Mutation.forHost { host = "gesg" } createResponseSelectionSet
            |> Graphql.Http.mutationRequest "http://localhost:4000/v2"


list : Cred -> BaseUrl -> Task Http.Error (List KnownHost)
list cred baseUrl =
    let
        endpoint =
            Endpoint.knownHosts (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl)
    in
        Decode.field "data" (Decode.list decoder)
            |> Api.getTask endpoint (Just cred)


create : Cred -> BaseUrl -> String -> (Result Http.Error KnownHost -> msg) -> Cmd msg
create cred baseUrl publicKey toMsg =
    let
        endpoint =
            Endpoint.knownHosts Nothing (Api.toEndpoint baseUrl)

        body =
            Encode.object [ ( "entry", Encode.string publicKey ) ]
                |> Http.jsonBody
    in
        Api.post endpoint (Just cred) body toMsg decoder
