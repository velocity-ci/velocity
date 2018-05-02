module Component.KnownHostForm
    exposing
        ( Context
        , Config
        , init
        , Field(..)
        , update
        , updateServerErrors
        , errorsDecoder
        , view
        , viewSubmitButton
        , submitting
        , submitValues
        , submit
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
import Page.Helpers exposing (ifBelowLength, ifAboveLength, validClasses, formatDateTime, sortByDatetime, getFieldErrors)
import Util exposing ((=>))
import Views.Form as Form
import Request.Errors


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


type alias Context =
    { form : KnownHostForm
    , errors : List Error
    , serverErrors : List Error
    , submitting : Bool
    }


newField : Field -> FormField
newField field =
    FormField "" False field


initialForm : KnownHostForm
initialForm =
    { scannedKey = newField ScannedKey }



-- UPDATE HELPERS --


resetServerErrors : List Error -> Field -> List Error
resetServerErrors errors field =
    let
        shouldInclude error =
            Tuple.first error /= field && Tuple.first error /= Form
    in
        List.filter shouldInclude errors


resetServerErrorsForField : Context -> Field -> List Error
resetServerErrorsForField context field =
    resetServerErrors context.serverErrors field


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


updateInput : Field -> String -> FormField
updateInput field value =
    FormField value True field


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


updateServerErrors : List ( String, String ) -> Context -> Context
updateServerErrors errorMessages context =
    { context | serverErrors = List.map serverErrorToFormError errorMessages }


submitting : Bool -> Context -> Context
submitting submitting context =
    { context | submitting = submitting }


submit : Context -> Context
submit context =
    { context
        | submitting = True
        , serverErrors = []
        , errors = []
    }


serverErrorToFormError : ( String, String ) -> Error
serverErrorToFormError ( fieldNameString, errorString ) =
    let
        field =
            case fieldNameString of
                "scanned_key" ->
                    ScannedKey

                _ ->
                    Form
    in
        ( field, errorString )


type alias Config msg =
    { setScannedKeyMsg : String -> msg
    , submitMsg : msg
    }


init : Context
init =
    { form = initialForm
    , errors = []
    , serverErrors = []
    , submitting = False
    }



-- VIEW --


view : Config msg -> Context -> Html msg
view { setScannedKeyMsg, submitMsg } context =
    let
        form =
            context.form

        combinedErrors =
            context.errors ++ context.serverErrors

        globalErrors =
            List.filter (\e -> (Tuple.first e) == Form) combinedErrors

        errors field =
            if field.dirty then
                getFieldErrors combinedErrors field
            else
                []

        inputClassList =
            validClasses combinedErrors

        scannedKeyField =
            Form.textarea
                { name = "scanned_key"
                , label = "Scanned key"
                , help = Nothing
                , errors = errors form.scannedKey
                }
                [ placeholder "ssh-keyscan <host>"
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
            (List.map viewGlobalError globalErrors
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
