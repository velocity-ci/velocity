module Page.KnownHosts exposing (..)

import Data.KnownHost as KnownHost exposing (KnownHost)
import Data.Session as Session exposing (Session)
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.KnownHost
import Views.Page as Page
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Util exposing ((=>))
import Views.Form as Form
import Validate exposing (..)
import Page.Helpers exposing (ifBelowLength, ifAboveLength, validClasses, formatDateTime, sortByDatetime, getFieldErrors)
import Route
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional)
import Json.Encode exposing (encode)


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


init : Session -> Task PageLoadError Model
init session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadKnownHosts =
            Request.KnownHost.list maybeAuthToken
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.KnownHosts "Known Hosts are currently unavailable."

        initialModel knownHosts =
            { formCollapsed = True
            , form = initialForm
            , errors = validate initialForm
            , serverErrors = []
            , submitting = False
            , knownHosts = knownHosts
            }
    in
        Task.map initialModel loadKnownHosts
            |> Task.mapError handleLoadError



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    div []
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
        div [ class "row" ]
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
                            "scannedKey"
                            "Scanned key"
                            (errors form.scannedKey)
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
        div [ class "row", style [ ( "margin-top", "3em" ) ] ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ] [ text ("KnownHosts (" ++ knownHostAmount ++ ")") ]
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
    | KnownHostCreated (Result Http.Error KnownHost)


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
                "scannedKey" ->
                    ScannedKey

                _ ->
                    Form
    in
        field => errorString


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
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
                                    |> Request.KnownHost.create submitValues
                                    |> Http.send KnownHostCreated

                            cmd =
                                session
                                    |> Session.attempt "create knownHost" cmdFromAuth
                                    |> Tuple.second
                        in
                            { model
                                | submitting = True
                                , serverErrors = []
                                , errors = []
                            }
                                => cmd

                    errors ->
                        { model | errors = errors }
                            => Cmd.none

            SetFormCollapsed state ->
                { model | formCollapsed = state }
                    => Cmd.none

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

            KnownHostCreated (Err err) ->
                let
                    errorMessages =
                        case err of
                            Http.BadStatus response ->
                                response.body
                                    |> decodeString (field "errors" errorsDecoder)
                                    |> Result.withDefault []

                            _ ->
                                [ ( "", "Unable to process knownHost." ) ]
                in
                    { model
                        | submitting = False
                        , serverErrors = List.map serverErrorToFormError errorMessages
                    }
                        => Cmd.none

            KnownHostCreated (Ok knownHost) ->
                { model
                    | knownHosts = knownHost :: model.knownHosts
                    , submitting = False
                    , formCollapsed = True
                    , form = initialForm
                }
                    => Cmd.none



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
        |> optionalError "key"


optionalError : String -> Decoder (List ( String, String ) -> a) -> Decoder a
optionalError fieldName =
    let
        errorToTuple errorMessage =
            ( fieldName, errorMessage )
    in
        optional fieldName (Decode.list (Decode.map errorToTuple string)) []
