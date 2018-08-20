module Component.KnownHostForm
    exposing
        ( Context
        , Config
        , init
        , Field(..)
        , update
        , errorsDecoder
        , view
        , viewSubmitButton
        , serverErrorToFormError
        , submitValues
        , isUntouched
        )

-- EXTERNAL

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Validate exposing (..)
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional)
import Bootstrap.Button as Button


-- INTERNAL

import Data.Project as Project exposing (Project)
import Data.GitUrl as GitUrl exposing (GitUrl)
import Util exposing ((=>))
import Views.Form as Form
import Component.Form as BaseForm exposing (..)


-- MODEL --


type Field
    = Form
    | ScannedKey


type alias KnownHostForm =
    { scannedKey : FormField Field
    }


type alias Context =
    BaseForm.Context Field KnownHostForm


type alias Error =
    BaseForm.Error Field


type alias Config msg =
    { gitUrl : Maybe GitUrl
    , setScannedKeyMsg : String -> msg
    , submitMsg : msg
    }


initialForm : KnownHostForm
initialForm =
    { scannedKey = newField ScannedKey }



-- UPDATE HELPERS --


updateForm : KnownHostForm -> Field -> String -> KnownHostForm
updateForm form field value =
    let
        updatedField =
            updateInput field value
    in
        case field of
            ScannedKey ->
                { form | scannedKey = updatedField }

            Form ->
                form


update : Context -> Field -> String -> Context
update context field value =
    let
        updatedForm =
            updateForm context.form field value
    in
        { context
            | errors = validate updatedForm
            , form = updatedForm
            , serverErrors = resetServerErrorsForField context field
        }


submitValues : Context -> { scannedKey : String }
submitValues { form } =
    { scannedKey = form.scannedKey.value }


serverErrorToFormError : ( String, String ) -> Error
serverErrorToFormError ( fieldNameString, errorString ) =
    let
        field =
            case fieldNameString of
                "scanned_key" ->
                    ScannedKey

                "entry" ->
                    ScannedKey

                _ ->
                    Form
    in
        ( field, errorString )


init : Context
init =
    { form = initialForm
    , errors = []
    , serverErrors = []
    , submitting = False
    }



-- VIEW --


view : Config msg -> Context -> Html msg
view { setScannedKeyMsg, submitMsg, gitUrl } context =
    let
        form =
            context.form

        combinedErrors =
            allErrors context

        errors field =
            if field.dirty then
                getFieldErrors combinedErrors field
            else
                []

        inputClassList =
            validClasses combinedErrors

        help =
            case gitUrl of
                Just { source } ->
                    String.join " "
                        [ "We require the host's public ssh key so we know which git servers to trust."
                        , "You can run `ssh-keyscan " ++ source ++ "` and verify the fingerprints are trusted for " ++ source ++ "."
                        , "e.g. GitHub publish theirs here: https://help.github.com/articles/github-s-ssh-key-fingerprints/"
                        ]

                Nothing ->
                    ""

        placeholderText =
            case gitUrl of
                Just { source } ->
                    "ssh-keyscan " ++ source

                Nothing ->
                    ""

        scannedKeyField =
            Form.textarea
                { name = "scanned_key"
                , label = "Scanned key"
                , help = Just help
                , errors = errors form.scannedKey
                }
                [ placeholder placeholderText
                , attribute "required" ""
                , rows 3
                , value form.scannedKey.value
                , onInput setScannedKeyMsg
                , classList <| inputClassList form.scannedKey
                , disabled context.submitting
                ]
                []
    in
        div []
            (List.map viewGlobalError (globalErrors Form combinedErrors)
                ++ [ Html.form [ attribute "novalidate" "", onSubmit submitMsg ]
                        [ scannedKeyField ]
                   ]
            )


viewGlobalError : Error -> Html msg
viewGlobalError error =
    div [ class "alert alert-danger" ] [ text (Tuple.second error) ]


isUntouched : Context -> Bool
isUntouched { form } =
    not form.scannedKey.dirty


viewSubmitButton : Config msg -> Context -> Html msg
viewSubmitButton { submitMsg } context =
    let
        hasErrors =
            not <| List.isEmpty context.errors

        submitting =
            context.submitting

        untouched =
            isUntouched context
    in
        Button.button
            [ Button.outlinePrimary
            , Button.attrs
                [ onClick submitMsg
                , disabled (hasErrors || submitting || untouched)
                ]
            ]
            [ text "Create" ]



-- VALIDATION --


validate : Validator ( Field, String ) KnownHostForm
validate =
    Validate.all
        [ (.scannedKey >> .value) >> (ifBelowLength 8) (ScannedKey => "Scanned key must be over 7 characters.")
        ]


errorsDecoder : Decoder (List ( String, String ))
errorsDecoder =
    decode (\scannedKey -> List.concat [ scannedKey ])
        |> optionalError "entry"
