port module Api
    exposing
        ( BaseUrl
        , Cred
        , application
        , get
        , getTask
        , login
        , logout
        , post
        , storeCredWith
        , toEndpoint
        , toWsEndpoint
        , username
        , signIn
        , viewerChanges
        , Response
        , responseMessages
        , responseResult
        , validationErrorSelectionSet
        , responseWasSuccessful
        , ValidationMessage
        , validationMessages
        )

{-| This module is responsible for communicating to the Conduit API.
It exposes an opaque Endpoint type which is guaranteed to point to the correct URL.
-}

import Api.Endpoint as Endpoint exposing (Endpoint)
import Browser
import Browser.Navigation as Nav
import Http exposing (Body, Expect)
import Json.Decode as Decode exposing (Decoder, Value, decodeString, field, string)
import Json.Decode.Pipeline exposing (required)
import Json.Encode as Encode
import Task exposing (Task)
import Url exposing (Url)
import Username exposing (Username)
import Api.Compiled.Object
import Api.Compiled.Query as Query
import Api.Compiled.Mutation as Mutation
import Graphql.Http
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import RemoteData
import Graphql.Operation exposing (RootQuery)
import Api.Compiled.Object.SessionPayload as SessionPayload
import Api.Compiled.Object.Session as Session
import Api.Compiled.Object.ValidationMessage as ValidationMessage
import Maybe.Extra


--import StarWars.Scalar exposing (Id(..))
--
-- CRED


{-| The base URL to use for all relative endpoints

This is just another endpoint which is good because it means only Endpoint can actually understand it

-}
type BaseUrl
    = BaseUrl Endpoint


toEndpoint : BaseUrl -> Endpoint
toEndpoint (BaseUrl endpoint) =
    endpoint


toWsEndpoint : BaseUrl -> String
toWsEndpoint (BaseUrl endpoint) =
    Endpoint.toWs endpoint


{-| The authentication credentials for the Viewer (that is, the currently logged-in user.)

This includes:

  - The cred's Username
  - The cred's authentication token

By design, there is no way to access the token directly as a String.
It can be encoded for persistence, and it can be added to a header
to a HttpBuilder for a request, but that's it.

This token should never be rendered to the end user, and with this API, it
can't be!

-}
type Cred
    = Cred Username String


username : Cred -> Username
username (Cred val _) =
    val


credHeader : Cred -> Http.Header
credHeader (Cred _ str) =
    Http.header "Authorization" ("Bearer " ++ str)


{-| It's important that this is never exposed!
We epxose `login` and `application` instead, so we can be certain that if anyone
ever has access to a `Cred` value, it came from either the login API endpoint
or was passed in via flags.
-}
credDecoder : Decoder Cred
credDecoder =
    Decode.succeed Cred
        |> required "username" Username.decoder
        |> required "token" Decode.string



-- PERSISTENCE


decode : Decoder (Cred -> viewer) -> Value -> Result Decode.Error viewer
decode decoder value =
    -- It's stored in localStorage as a JSON String;
    -- first decode the Value as a String, then
    -- decode that String as JSON.
    Decode.decodeValue Decode.string value
        |> Result.andThen (\str -> Decode.decodeString (Decode.field "user" (decoderFromCred decoder)) str)


port onStoreChange : (Value -> msg) -> Sub msg


viewerChanges : (Maybe viewer -> msg) -> Decoder (Cred -> viewer) -> Sub msg
viewerChanges toMsg decoder =
    onStoreChange (\value -> toMsg (decodeFromChange decoder value))


decodeFromChange : Decoder (Cred -> viewer) -> Value -> Maybe viewer
decodeFromChange viewerDecoder val =
    -- It's stored in localStorage as a JSON String;
    -- first decode the Value as a String, then
    -- decode that String as JSON.
    Decode.decodeValue (storageDecoder viewerDecoder) val
        |> Result.toMaybe


storeCredWith : Cred -> Cmd msg
storeCredWith (Cred uname token) =
    let
        json =
            Encode.object
                [ ( "user"
                  , Encode.object
                        [ ( "username", Username.encode uname )
                        , ( "token", Encode.string token )
                        ]
                  )
                ]
    in
        storeCache (Just json)


logout : Cmd msg
logout =
    storeCache Nothing


port storeCache : Maybe Value -> Cmd msg



-- SERIALIZATION
-- APPLICATION


application :
    Decoder (Cred -> viewer)
    -> (BaseUrl -> { width : Int, height : Int } -> context)
    -> { init : Maybe viewer -> Result Decode.Error context -> Url -> Nav.Key -> ( model, Cmd msg )
       , onUrlChange : Url -> msg
       , onUrlRequest : Browser.UrlRequest -> msg
       , subscriptions : model -> Sub msg
       , update : msg -> model -> ( model, Cmd msg )
       , view : model -> Browser.Document msg
       }
    -> Program Value model msg
application viewerDecoder toContext config =
    let
        init flags url navKey =
            let
                maybeViewer =
                    Decode.decodeValue (Decode.field "viewer" Decode.string) flags
                        |> Result.andThen (Decode.decodeString (storageDecoder viewerDecoder))
                        |> Result.toMaybe

                baseUrlResult =
                    Decode.decodeValue (Decode.field "baseUrl" Decode.string) flags
                        |> Result.map (Endpoint.fromString >> BaseUrl)

                deviceDimensionResult =
                    Decode.decodeValue
                        (Decode.map2 (\width height -> { width = width, height = height })
                            (Decode.field "width" Decode.int)
                            (Decode.field "height" Decode.int)
                        )
                        flags

                contextResult =
                    Result.map2 toContext baseUrlResult deviceDimensionResult
            in
                config.init maybeViewer contextResult url navKey
    in
        Browser.application
            { init = init
            , onUrlChange = config.onUrlChange
            , onUrlRequest = config.onUrlRequest
            , subscriptions = config.subscriptions
            , update = config.update
            , view = config.view
            }


storageDecoder : Decoder (Cred -> viewer) -> Decoder viewer
storageDecoder viewerDecoder =
    Decode.field "user" (decoderFromCred viewerDecoder)



-- HTTP


get : Endpoint -> Maybe Cred -> (Result Http.Error a -> msg) -> Decoder a -> Cmd msg
get url maybeCred toMsg decoder =
    Endpoint.request
        { method = "GET"
        , url = url
        , expect = Http.expectJson toMsg decoder
        , headers =
            case maybeCred of
                Just cred ->
                    [ credHeader cred ]

                Nothing ->
                    []
        , body = Http.emptyBody
        , timeout = Nothing
        , withCredentials = False
        }


getTask : Endpoint -> Maybe Cred -> Decoder a -> Task Http.Error a
getTask url maybeCred decoder =
    Endpoint.task
        { method = "GET"
        , url = url
        , resolver =
            Http.stringResolver
                (\res ->
                    case res of
                        Http.BadUrl_ url_ ->
                            Err (Http.BadUrl url_)

                        Http.Timeout_ ->
                            Err Http.Timeout

                        Http.NetworkError_ ->
                            Err Http.NetworkError

                        Http.BadStatus_ metadata body ->
                            Err (Http.BadStatus metadata.statusCode)

                        Http.GoodStatus_ _ body ->
                            case Decode.decodeString decoder body of
                                Ok value ->
                                    Ok value

                                Err err ->
                                    Err (Http.BadBody (Decode.errorToString err))
                )
        , headers =
            case maybeCred of
                Just cred ->
                    [ credHeader cred ]

                Nothing ->
                    []
        , body = Http.emptyBody
        , timeout = Nothing
        , withCredentials = False
        }


put : Endpoint -> Cred -> Body -> (Result Http.Error a -> msg) -> Decoder a -> Cmd msg
put url cred body toMsg decoder =
    Endpoint.request
        { method = "PUT"
        , url = url
        , expect = Http.expectJson toMsg decoder
        , headers = [ credHeader cred ]
        , body = body
        , timeout = Nothing
        , withCredentials = False
        }


post : Endpoint -> Maybe Cred -> Body -> (Result Http.Error a -> msg) -> Decoder a -> Cmd msg
post url maybeCred body toMsg decoder =
    Endpoint.request
        { method = "POST"
        , url = url
        , expect = Http.expectJson toMsg decoder
        , headers =
            case maybeCred of
                Just cred ->
                    [ credHeader cred ]

                Nothing ->
                    []
        , body = body
        , timeout = Nothing
        , withCredentials = False
        }


delete : Endpoint -> Cred -> Body -> (Result Http.Error a -> msg) -> Decoder a -> Cmd msg
delete url cred body toMsg decoder =
    Endpoint.request
        { method = "DELETE"
        , url = url
        , expect = Http.expectJson toMsg decoder
        , headers = [ credHeader cred ]
        , body = body
        , timeout = Nothing
        , withCredentials = False
        }



--
--


login : BaseUrl -> Http.Body -> Decoder (Cred -> a) -> (Result Http.Error a -> msg) -> Cmd msg
login (BaseUrl baseUrl) body decoder toMsg =
    post (Endpoint.login baseUrl) Nothing body toMsg (decoderFromCred decoder)



--query : SelectionSet Response RootQuery
--query =
--    SelectionSet.map Response Query.hello
--sessionSelectionSet : SelectionSet String Api.Compiled.Object.SessionPayload
--sessionSelectionSet =
--    SelectionSet.map identity
--query : SelectionSet Response SessionPayload
--query =
--    SelectionSet.map Response Query.hello
--


type ValidationMessage
    = ValidationMessage ValidationMessageRec


type alias ValidationMessageRec =
    { field : Maybe String
    , message : Maybe String
    }


type Response result
    = Response (ResponseRec result)


type alias ResponseRec result =
    { successful : Bool
    , messages : Maybe (List (Maybe ValidationMessage))
    , result : Maybe result
    }


validationMessages : List ValidationMessage -> List ( String, String )
validationMessages =
    List.filterMap (\(ValidationMessage { field, message }) -> Maybe.map2 Tuple.pair field message)


responseResult : Response a -> Maybe a
responseResult (Response { result }) =
    result


type alias Message =
    { field : String
    , message : String
    }


responseMessages : Response a -> List Message
responseMessages (Response { messages }) =
    messages
        |> Maybe.map (List.filterMap identity)
        |> Maybe.withDefault []
        |> List.filterMap (\(ValidationMessage m) -> Maybe.map2 Message m.field m.message)


responseWasSuccessful : Response a -> Bool
responseWasSuccessful (Response { successful }) =
    successful


validationErrorSelectionSet : SelectionSet ValidationMessage Api.Compiled.Object.ValidationMessage
validationErrorSelectionSet =
    SelectionSet.map2 (\field message -> ValidationMessage (ValidationMessageRec field message))
        ValidationMessage.field
        ValidationMessage.message



--
--
--responseSelectionSet =
--    SelectionSet.succeed Response
--        |> with internalSelectionSet


signIn : BaseUrl -> Mutation.SignInRequiredArguments -> Graphql.Http.Request (Maybe (Response Cred))
signIn (BaseUrl baseUrl) values =
    let
        usernameSelectionSet =
            SelectionSet.map (Maybe.map Username.fromString) (SessionPayload.result Session.username)

        viewerSelectionSet =
            SelectionSet.map2 Cred
                (SelectionSet.nonNullOrFail usernameSelectionSet)
                (SelectionSet.nonNullOrFail <| SessionPayload.result Session.token)

        selectionSet =
            SelectionSet.map3 (\successful messages result -> Response (ResponseRec successful messages result))
                SessionPayload.successful
                (SessionPayload.messages (validationErrorSelectionSet))
                (SelectionSet.map Just viewerSelectionSet)
    in
        Mutation.signIn values selectionSet
            |> Graphql.Http.mutationRequest "http://localhost:4000/v2"



--
--createKnownHost : BaseUrl -> Mutation.ForHostRequiredArguments -> Graphql.Http.Request (Maybe (Response KnownHost))
--createKnownHost (BaseUrl baseUrl) values =
--
--signIn : BaseUrl -> Mutation.SignInRequiredArguments -> Decoder (Cred -> a) -> (Result Http.Error a -> msg) -> Cmd msg
--signIn (BaseUrl baseUrl) body decoder toMsg =
--    Mutation.signIn body (SelectionSet.map identity)
--        |> Graphql.Http.mutationRequest "test"
--        |> Graphql.Http.send (RemoteData.fromResult >> toMsg)
--
--
--
--
--register : Http.Body -> Decoder (Cred -> a) -> Http.Request a
--register body decoder =
--    post Endpoint.users Nothing body (Decode.field "user" (decoderFromCred decoder))
--
--
--settings : Cred -> Http.Body -> Decoder (Cred -> a) -> Http.Request a
--settings cred body decoder =
--    put Endpoint.user cred body (Decode.field "user" (decoderFromCred decoder))
--


decoderFromCred : Decoder (Cred -> a) -> Decoder a
decoderFromCred decoder =
    Decode.map2 (\fromCred cred -> fromCred cred)
        decoder
        credDecoder



-- ERRORS


addServerError : List String -> List String
addServerError list =
    "Server error" :: list


{-| Many API endpoints include an "errors" field in their BadStatus responses.
-}



--decodeErrors : Http.Response a -> List String
--decodeErrors response =
--    case response of
--        Http.BadStatus response ->
--            response.body
--                |> decodeString (field "errors" errorsDecoder)
--                |> Result.withDefault [ "Server error" ]
--
--        _ ->
--            [ "Server error" ]


errorsDecoder : Decoder (List String)
errorsDecoder =
    Decode.keyValuePairs (Decode.list Decode.string)
        |> Decode.map (List.concatMap fromPair)


fromPair : ( String, List String ) -> List String
fromPair ( field, errors ) =
    List.map (\error -> field ++ " " ++ error) errors



-- LOCALSTORAGE KEYS


cacheStorageKey : String
cacheStorageKey =
    "cache"


credStorageKey : String
credStorageKey =
    "cred"
