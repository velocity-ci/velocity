module KnownHost
    exposing
        ( KnownHost
        , CreateResponse(..)
        , addKnownHost
        , findKnownHost
        , isUnknownHost
        , findForGitUrl
        , isVerified
        , md5
        , sha256
        , selectionSet
        , createUnverified
        , verify
        )

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
    , verified : Bool
    }



-- SERIALIZATION --


internalSelectionSet : SelectionSet Internals Api.Compiled.Object.KnownHost
internalSelectionSet =
    SelectionSet.succeed Internals
        |> with idSelectionSet
        |> with KnownHost.host
        |> with KnownHost.fingerprintMd5
        |> with KnownHost.fingerprintSha256
        |> with KnownHost.verified


selectionSet : SelectionSet KnownHost Api.Compiled.Object.KnownHost
selectionSet =
    SelectionSet.succeed KnownHost
        |> with internalSelectionSet



-- INFO


md5 : KnownHost -> String
md5 (KnownHost knownHost) =
    knownHost.md5


sha256 : KnownHost -> String
sha256 (KnownHost knownHost) =
    knownHost.sha256


host : KnownHost -> String
host (KnownHost knownHost) =
    knownHost.host


isVerified : KnownHost -> Bool
isVerified (KnownHost knownHost) =
    knownHost.verified



-- HELPERS --


findForGitUrl : List KnownHost -> GitUrl -> Maybe KnownHost
findForGitUrl knownHosts { source } =
    knownHosts
        |> List.filter (\(KnownHost a) -> a.host == source)
        |> List.head


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
    List.map host knownHosts


findKnownHost : List KnownHost -> KnownHost -> Maybe KnownHost
findKnownHost knownHosts (KnownHost knownHost) =
    List.filter (\(KnownHost a) -> a.id == knownHost.id) knownHosts
        |> List.head


addKnownHost : List KnownHost -> KnownHost -> List KnownHost
addKnownHost knownHosts knownHost =
    case findKnownHost knownHosts knownHost of
        Just _ ->
            updateKnownHost knownHosts knownHost

        Nothing ->
            knownHost :: knownHosts


updateKnownHost : List KnownHost -> KnownHost -> List KnownHost
updateKnownHost knownHosts (KnownHost b) =
    List.map
        (\(KnownHost a) ->
            if a.id == b.id then
                KnownHost b
            else
                KnownHost a
        )
        knownHosts



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


createUnverified : Cred -> BaseUrl -> Mutation.ForHostRequiredArguments -> Graphql.Http.Request (Maybe CreateResponse)
createUnverified cred baseUrl values =
    let
        endpoint =
            Api.toEndpoint baseUrl
                |> Endpoint.unwrap
    in
        Mutation.forHost values createResponseSelectionSet
            |> Graphql.Http.mutationRequest "http://localhost:4000/v2"


verify : Cred -> BaseUrl -> KnownHost -> Graphql.Http.Request (Maybe KnownHost)
verify cred baseUrl (KnownHost { id }) =
    let
        endpoint =
            Api.toEndpoint baseUrl
                |> Endpoint.unwrap
    in
        Mutation.verify { id = idToString id } selectionSet
            |> Graphql.Http.mutationRequest "http://localhost:4000/v2"
