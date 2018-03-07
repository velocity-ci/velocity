module Page.KnownHosts exposing (..)

import Context exposing (Context)
import Data.KnownHost as KnownHost exposing (KnownHost)
import Data.Session as Session exposing (Session)
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.KnownHost
import Request.Errors
import Views.Page as Page
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Util exposing ((=>))
import Views.Form as Form
import Validate exposing (..)
import Page.Helpers exposing (ifBelowLength, ifAboveLength, validClasses, formatDateTime, sortByDatetime, getFieldErrors)
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional)


-- MODEL --


type Field
    = Form
    | ScannedKey


type alias FormField =
    { value : String
    , dirty : Bool
    , field : Field
    }


type alias KnownHostForm =
    { scannedKey : FormField
    }


type alias Model =
    { formCollapsed : Bool
    , form : KnownHostForm
    , errors : List Error
    , serverErrors : List Error
    , submitting : Bool
    , knownHosts : List KnownHost
    }


newField : Field -> FormField
newField field =
    FormField "" False field


initialForm : KnownHostForm
initialForm =
    { scannedKey = newField ScannedKey
    }


init : Context -> Session msg -> Task (Request.Errors.Error PageLoadError) Model
init context session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadKnownHosts =
            Request.KnownHost.list context maybeAuthToken

        loadError =
            pageLoadError Page.KnownHosts "Known hosts are currently unavailable."

        initialModel (Paginated { total, results }) =
            { formCollapsed = True
            , form = initialForm
            , errors = validate initialForm
            , serverErrors = []
            , submitting = False
            , knownHosts = results
            }
    in
        Task.map initialModel loadKnownHosts
            |> Task.mapError (Request.Errors.withDefaultError loadError)



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    div [ class "container-fluid" ]
        [ viewKnownHostFormContainer model
        , viewKnownHostList model.knownHosts
        ]


viewKnownHostFormContainer : Model -> Html Msg
viewKnownHostFormContainer model =
    let
        toggleClassList =
            [ ( "fa-plus", model.formCollapsed )
            , ( "fa-minus", not model.formCollapsed )
            ]
    in
        div [ class "row default-margin-top" ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ]
                        [ text "Add Known Host"
                        , button
                            [ type_ "button"
                            , class "btn btn-primary btn-sm float-right"
                            , onClick (SetFormCollapsed <| not model.formCollapsed)
                            , disabled model.submitting
                            ]
                            [ i [ class "fa", classList toggleClassList ] [] ]
                        ]
                    , Util.viewIf (not model.formCollapsed) <| viewKnownHostForm model
                    ]
                ]
            ]


viewGlobalError : Error -> Html Msg
viewGlobalError error =
    div [ class "alert alert-danger" ] [ text (Tuple.second error) ]


viewKnownHostForm : Model -> Html Msg
viewKnownHostForm model =
    let
        form =
            model.form

        combinedErrors =
            model.errors ++ model.serverErrors

        inputClassList =
            validClasses combinedErrors

        errors field =
            if field.dirty then
                getFieldErrors combinedErrors field
            else
                []

        globalErrors =
            List.filter (\e -> (Tuple.first e) == Form) combinedErrors
    in
        div [ class "card-body" ]
            (List.map viewGlobalError globalErrors
                ++ [ Html.form [ attribute "novalidate" "", onSubmit SubmitForm ]
                        [ Form.textarea
                            { name = "scanned_key"
                            , label = "Scanned key"
                            , help = Nothing
                            , errors = errors form.scannedKey
                            }
                            [ placeholder "Scanned key (ssh-keyscan <host>)"
                            , attribute "required" ""
                            , rows 3
                            , value form.scannedKey.value
                            , onInput SetScannedKey
                            , classList <| inputClassList form.scannedKey
                            , disabled model.submitting
                            ]
                            []
                        , button
                            [ class "btn btn-primary"
                            , type_ "submit"
                            , disabled ((not <| List.isEmpty combinedErrors) || model.submitting)
                            ]
                            [ text "Submit" ]
                        , Util.viewIf model.submitting Form.viewSpinner
                        ]
                   ]
            )


viewKnownHostList : List KnownHost -> Html Msg
viewKnownHostList knownHosts =
    let
        knownHostAmount =
            knownHosts
                |> List.length
                |> toString
    in
        div [ class "row default-margin-top" ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ] [ text ("Known hosts (" ++ knownHostAmount ++ ")") ]
                    , ul [ class "list-group" ] (List.map viewKnownHostListItem knownHosts)
                    ]
                ]
            ]


viewKnownHostListItem : KnownHost -> Html Msg
viewKnownHostListItem knownHost =
    li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h5 [ class "mb-1" ] [ text (String.join "," knownHost.hosts) ] ]
        , small []
            [ text knownHost.sha256 ]
        ]



-- UPDATE --


type Msg
    = SubmitForm
    | SetFormCollapsed Bool
    | SetScannedKey String
    | KnownHostCreated (Result Request.Errors.HttpError KnownHost)


type ExternalMsg
    = NoOp
    | HandleRequestError Request.Errors.HandledError


updateInput : Field -> String -> FormField
updateInput field value =
    FormField value True field


resetServerErrors : List Error -> Field -> List Error
resetServerErrors errors field =
    let
        shouldInclude error =
            let
                errorField =
                    Tuple.first error
            in
                errorField /= field && errorField /= Form
    in
        List.filter shouldInclude errors


serverErrorToFormError : ( String, String ) -> Error
serverErrorToFormError ( fieldNameString, errorString ) =
    let
        field =
            case fieldNameString of
                "entry" ->
                    ScannedKey

                _ ->
                    Form
    in
        field => errorString


update : Context -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context session msg model =
    let
        form =
            model.form

        resetServerErrorsForField =
            resetServerErrors model.serverErrors
    in
        case msg of
            SubmitForm ->
                case validate form of
                    [] ->
                        let
                            submitValues =
                                { scannedKey = form.scannedKey.value
                                }

                            cmdFromAuth authToken =
                                authToken
                                    |> Request.KnownHost.create context submitValues
                                    |> Task.attempt KnownHostCreated

                            cmd =
                                session
                                    |> Session.attempt "create known host" cmdFromAuth
                                    |> Tuple.second
                        in
                            { model
                                | submitting = True
                                , serverErrors = []
                                , errors = []
                            }
                                => cmd
                                => NoOp

                    errors ->
                        { model | errors = errors }
                            => Cmd.none
                            => NoOp

            SetFormCollapsed state ->
                { model | formCollapsed = state }
                    => Cmd.none
                    => NoOp

            SetScannedKey scannedKey ->
                let
                    newForm =
                        { form | scannedKey = scannedKey |> (updateInput ScannedKey) }
                in
                    { model
                        | errors = validate newForm
                        , form = newForm
                        , serverErrors = resetServerErrorsForField ScannedKey
                    }
                        => Cmd.none
                        => NoOp

            KnownHostCreated (Err err) ->
                let
                    newState =
                        { model | submitting = False }

                    stateWithErrorMessages errorMessages =
                        { newState
                            | serverErrors = List.map serverErrorToFormError errorMessages
                        }
                            => Cmd.none
                            => NoOp
                in
                    case err of
                        Request.Errors.HandledError handledError ->
                            newState => Cmd.none => HandleRequestError handledError

                        Request.Errors.UnhandledError (Http.BadStatus response) ->
                            response.body
                                |> decodeString errorsDecoder
                                |> Result.withDefault []
                                |> stateWithErrorMessages

                        _ ->
                            stateWithErrorMessages [ ( "", "Unable to process knownHost." ) ]

            KnownHostCreated (Ok knownHost) ->
                { model
                    | knownHosts = knownHost :: model.knownHosts
                    , submitting = False
                    , formCollapsed = True
                    , form = initialForm
                }
                    => Cmd.none
                    => NoOp



-- VALIDATION --


type alias Error =
    ( Field, String )


validate : Validator ( Field, String ) KnownHostForm
validate =
    Validate.all
        [ (.scannedKey >> .value) >> (ifBelowLength 8) (ScannedKey => "Private key must be over 7 characters.")
        ]


errorsDecoder : Decoder (List ( String, String ))
errorsDecoder =
    decode (\scannedKey -> List.concat [ scannedKey ])
        |> optionalError "entry"


optionalError : String -> Decoder (List ( String, String ) -> a) -> Decoder a
optionalError fieldName =
    let
        errorToTuple errorMessage =
            ( fieldName, errorMessage )
    in
        optional fieldName (Decode.list (Decode.map errorToTuple string)) []
