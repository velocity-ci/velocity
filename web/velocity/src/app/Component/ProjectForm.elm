module Component.ProjectForm
    exposing
        ( Context
        , Config
        , Field(..)
        , init
        , view
        , viewSubmitButton
        , update
        , validate
        , submitValues
        , errorsDecoder
        , serverErrorToFormError
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
import Page.Helpers exposing (formatDateTime, sortByDatetime)
import Util exposing ((=>))
import Views.Form as Form
import Request.Errors
import Component.Form as BaseForm exposing (..)


-- MODEL --


type Field
    = Form
    | Name
    | Repository
    | PrivateKey


type alias ProjectForm =
    { name : FormField Field
    , repository : FormField Field
    , privateKey : FormField Field
    }


initialForm : ProjectForm
initialForm =
    { name = newField Name
    , repository = newField Repository
    , privateKey = newField PrivateKey
    }


type alias Config msg =
    { setNameMsg : String -> msg
    , setRepositoryMsg : String -> msg
    , setPrivateKeyMsg : String -> msg
    , submitMsg : msg
    }


type alias Context =
    BaseForm.Context Field ProjectForm


init : Context
init =
    { form = initialForm
    , errors = []
    , serverErrors = []
    , submitting = False
    }



-- UPDATE HELPERS --


updateForm : ProjectForm -> Field -> String -> ProjectForm
updateForm form field value =
    let
        updatedField =
            updateInput field value
    in
        case field of
            Repository ->
                { form | repository = updatedField }

            Name ->
                { form | name = updatedField }

            PrivateKey ->
                { form | privateKey = updatedField }

            Form ->
                form


serverErrorToFormError : ( String, String ) -> Error Field
serverErrorToFormError ( fieldNameString, errorString ) =
    let
        field =
            case fieldNameString of
                "name" ->
                    Name

                "repository" ->
                    Repository

                "key" ->
                    PrivateKey

                _ ->
                    Form
    in
        ( field, errorString )


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


isUntouched : Context -> Bool
isUntouched { form } =
    [ form.name.dirty, form.repository.dirty, form.privateKey.dirty ]
        |> List.member True
        |> not


submitValues : Context -> { name : String, repository : String, privateKey : Maybe String }
submitValues { form } =
    let
        privateKey =
            if isSshAddress form.repository.value then
                Just form.privateKey.value
            else
                Nothing
    in
        { name = form.name.value
        , repository = form.repository.value
        , privateKey = privateKey
        }



-- VIEW --


view : Config msg -> Context -> Html msg
view { setNameMsg, setRepositoryMsg, setPrivateKeyMsg, submitMsg } context =
    let
        form =
            context.form

        publicRepository =
            not (isSshAddress form.repository.value)

        combinedErrors =
            context.errors ++ context.serverErrors

        inputClassList =
            validClasses combinedErrors

        globalErrors =
            List.filter (\e -> (Tuple.first e) == Form) combinedErrors

        errors field =
            if field.dirty then
                getFieldErrors combinedErrors field
            else
                []

        nameField =
            Form.input
                { name = "name"
                , label = "Name"
                , help = Nothing
                , errors = errors form.name
                }
                [ attribute "required" ""
                , value form.name.value
                , onInput setNameMsg
                , classList <| inputClassList form.name
                , disabled context.submitting
                ]
                []

        repositoryField =
            Form.input
                { name = "repository"
                , label = "Repository address"
                , help = Just "Use a GIT+SSH address for authenticated repositories, otherwise use a HTTP(S) address."
                , errors = errors form.repository
                }
                [ attribute "required" ""
                , value form.repository.value
                , onInput setRepositoryMsg
                , classList <| inputClassList form.repository
                , disabled context.submitting
                ]
                []

        privateKeyField =
            let
                help =
                    if publicRepository then
                        "Not required for HTTP(S) repositories."
                    else
                        "The private key required to access this repository."
            in
                Form.textarea
                    { name = "key"
                    , label = "Private key"
                    , help = Just help
                    , errors = errors form.privateKey
                    }
                    [ attribute "required" ""
                    , rows 5
                    , value form.privateKey.value
                    , onInput setPrivateKeyMsg
                    , classList <| inputClassList form.privateKey
                    , disabled (context.submitting || publicRepository)
                    ]
                    []
    in
        div []
            (List.map viewGlobalError globalErrors
                ++ [ Html.form [ attribute "novalidate" "", onSubmit submitMsg ]
                        [ nameField
                        , repositoryField
                        , Util.viewIf (not publicRepository) privateKeyField
                        ]
                   ]
            )


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


viewGlobalError : Error Field -> Html msg
viewGlobalError error =
    div [ class "alert alert-danger" ] [ text (Tuple.second error) ]



-- VALIDATION --


isSshAddress : String -> Bool
isSshAddress address =
    String.slice 0 3 address == "git"


validate : Validator (Error Field) ProjectForm
validate =
    let
        privateKeyValidator ( privateKey, repository ) =
            if isSshAddress repository.value then
                String.length privateKey.value < 8
            else
                False
    in
        Validate.all
            [ (.name >> .value) >> (ifBelowLength 3) (Name => "Name must be over 2 characters.")
            , (.name >> .value) >> (ifAboveLength 128) (Name => "Name must be less than 129 characters.")
            , (.repository >> .value) >> (ifBelowLength 8) (Repository => "Repository must be over 7 characters.")
            , (.repository >> .value) >> (ifAboveLength 128) (Repository => "Repository must less than 129 characters.")
            , (\f -> ( f.privateKey, f.repository )) >> (ifInvalid privateKeyValidator) (PrivateKey => "Private key must be over 7 characters.")
            ]


errorsDecoder : Decoder (List ( String, String ))
errorsDecoder =
    decode (\name repository privateKey -> List.concat [ name, repository, privateKey ])
        |> optionalError "name"
        |> optionalError "repository"
        |> optionalError "key"
