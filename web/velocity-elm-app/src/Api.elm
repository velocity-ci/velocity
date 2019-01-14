port module Api
    exposing
        ( BaseUrl
        , Cred
        , application
        , logout
        , storeCredWith
        , username
        , signIn
        , viewerChanges
        , Response
        , responseMessages
        , queryRequest
        , responseResult
        , validationErrorSelectionSet
        , responseWasSuccessful
        , mutationRequest
        , ValidationMessage
        , validationMessages
        )

{-| This module is responsible for communicating to the Conduit API.
It exposes an opaque Endpoint type which is guaranteed to point to the correct URL.
-}

import Browser
import Browser.Navigation as Nav
import Http exposing (Body, Expect)
import Json.Decode as Decode exposing (Decoder, Value, field)
import Json.Decode.Pipeline exposing (required)
import Json.Encode as Encode
import Url exposing (Url)
import Username exposing (Username)
import Api.Compiled.Object
import Api.Compiled.Mutation as Mutation
import Graphql.Http exposing (Request)
import Graphql.Operation exposing (RootMutation, RootQuery)
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Api.Compiled.Object.SessionPayload as SessionPayload
import Api.Compiled.Object.Session as Session
import Api.Compiled.Object.ValidationMessage as ValidationMessage


-- CRED


{-| The base URL to use for all relative endpoints

This is just another endpoint which is good because it means only Endpoint can actually understand it

-}
type BaseUrl
    = BaseUrl String


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



-- REQUESTS


mutationRequest : BaseUrl -> SelectionSet decodesTo RootMutation -> Request decodesTo
mutationRequest (BaseUrl baseUrl) mutationSelectionSet =
    Graphql.Http.mutationRequest baseUrl mutationSelectionSet


queryRequest : BaseUrl -> SelectionSet decodesTo RootQuery -> Request decodesTo
queryRequest (BaseUrl baseUrl) querySelectionSet =
    Graphql.Http.queryRequest baseUrl querySelectionSet



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
                        |> Result.map BaseUrl

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


signIn : BaseUrl -> Mutation.SignInRequiredArguments -> Graphql.Http.Request (Response Cred)
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
