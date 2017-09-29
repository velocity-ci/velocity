module Page.Projects exposing (..)

import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Task exposing (Task)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
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
import Page.Project.Route as ProjectRoute


-- MODEL --


type Field
    = Form
    | Name
    | Repository
    | PrivateKey


type alias FormField =
    { value : String
    , dirty : Bool
    , field : Field
    }


type alias ProjectForm =
    { name : FormField
    , repository : FormField
    , privateKey : FormField
    }


type alias Model =
    { formCollapsed : Bool
    , form : ProjectForm
    , errors : List Error
    , serverErrors : List Error
    , submitting : Bool
    , projects : List Project
    }


newField : Field -> FormField
newField field =
    FormField "" False field


initialForm : ProjectForm
initialForm =
    { name = newField Name
    , repository = newField Repository
    , privateKey = newField PrivateKey
    }


init : Session -> Task PageLoadError Model
init session =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProjects =
            Request.Project.list maybeAuthToken
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Projects "Projects are currently unavailable."

        initialModel projects =
            { formCollapsed = True
            , form = initialForm
            , errors = validate initialForm
            , serverErrors = []
            , submitting = False
            , projects = projects
            }
    in
        Task.map initialModel loadProjects
            |> Task.mapError handleLoadError



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    div []
        [ div [ class "container-fluid" ]
            [ viewProjectFormContainer model
            , viewProjectList model.projects
            ]
        ]


viewBreadcrumb : Html Msg
viewBreadcrumb =
    div [ class "d-flex justify-content-start align-items-center bg-dark", style [ ( "height", "50px" ) ] ]
        [ div [ class "p-2" ]
            [ ol [ class "breadcrumb bg-dark", style [ ( "margin", "0" ) ] ]
                [ li [ class "breadcrumb-item active" ] [ text "Projects" ]
                ]
            ]
        ]


viewProjectFormContainer : Model -> Html Msg
viewProjectFormContainer model =
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
                        [ text "Create Project"
                        , button
                            [ type_ "button"
                            , class "btn btn-primary btn-sm float-right"
                            , onClick (SetFormCollapsed <| not model.formCollapsed)
                            , disabled model.submitting
                            ]
                            [ i [ class "fa", classList toggleClassList ] [] ]
                        ]
                    , Util.viewIf (not model.formCollapsed) <| viewProjectForm model
                    ]
                ]
            ]


viewGlobalError : Error -> Html Msg
viewGlobalError error =
    div [ class "alert alert-danger" ] [ text (Tuple.second error) ]


viewProjectForm : Model -> Html Msg
viewProjectForm model =
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
                        [ Form.input
                            "name"
                            "Name"
                            (errors form.name)
                            [ placeholder "Name"
                            , attribute "required" ""
                            , value form.name.value
                            , onInput SetName
                            , classList <| inputClassList form.name
                            , disabled model.submitting
                            ]
                            []
                        , Form.input
                            "repository"
                            "Repository address"
                            (errors form.repository)
                            [ placeholder "Repository"
                            , attribute "required" ""
                            , value form.repository.value
                            , onInput SetRepository
                            , classList <| inputClassList form.repository
                            , disabled model.submitting
                            ]
                            []
                        , Form.textarea
                            "key"
                            "Private key"
                            (errors form.privateKey)
                            [ placeholder "Private key"
                            , attribute "required" ""
                            , rows 3
                            , value form.privateKey.value
                            , onInput SetPrivateKey
                            , classList <| inputClassList form.privateKey
                            , disabled model.submitting
                            ]
                            []
                        , button
                            [ class "btn btn-primary"
                            , type_ "submit"
                            , disabled ((not <| List.isEmpty model.errors) || model.submitting)
                            ]
                            [ text "Submit" ]
                        , Util.viewIf model.submitting Form.viewSpinner
                        ]
                   ]
            )


viewProjectList : List Project -> Html Msg
viewProjectList projects =
    let
        projectAmount =
            projects
                |> List.length
                |> toString

        latestProjects =
            sortByDatetime .updatedAt projects
    in
        div [ class "row default-margin-top" ]
            [ div [ class "col-12" ]
                [ div [ class "card" ]
                    [ h4 [ class "card-header" ] [ text ("Projects (" ++ projectAmount ++ ")") ]
                    , ul [ class "list-group" ] (List.map viewProjectListItem latestProjects)
                    ]
                ]
            ]


viewProjectListItem : Project -> Html Msg
viewProjectListItem project =
    li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
        [ div [ class "d-flex w-100 justify-content-between" ]
            [ h5 [ class "mb-1" ]
                [ a [ Route.href (Route.Project ProjectRoute.Overview project.id) ] [ text project.name ] ]
            , small []
                [ text (formatDateTime project.updatedAt) ]
            ]
        , small []
            [ text project.repository ]
        ]



-- UPDATE --


type Msg
    = SubmitForm
    | SetFormCollapsed Bool
    | SetName String
    | SetRepository String
    | SetPrivateKey String
    | ProjectCreated (Result Http.Error Project)


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
                "name" ->
                    Name

                "repository" ->
                    Repository

                "key" ->
                    PrivateKey

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
                                { name = form.name.value
                                , repository = form.repository.value
                                , privateKey = form.privateKey.value
                                }

                            cmdFromAuth authToken =
                                authToken
                                    |> Request.Project.create submitValues
                                    |> Http.send ProjectCreated

                            cmd =
                                session
                                    |> Session.attempt "create project" cmdFromAuth
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

            SetName name ->
                let
                    newForm =
                        { form | name = name |> (updateInput Name) }
                in
                    { model
                        | errors = validate newForm
                        , form = newForm
                        , serverErrors = resetServerErrorsForField Name
                    }
                        => Cmd.none

            SetRepository repository ->
                let
                    newForm =
                        { form | repository = repository |> (updateInput Repository) }
                in
                    { model
                        | errors = validate newForm
                        , form = newForm
                        , serverErrors = resetServerErrorsForField Repository
                    }
                        => Cmd.none

            SetPrivateKey privateKey ->
                let
                    newForm =
                        { form | privateKey = privateKey |> (updateInput PrivateKey) }
                in
                    { model
                        | errors = validate newForm
                        , form = newForm
                        , serverErrors = resetServerErrorsForField PrivateKey
                    }
                        => Cmd.none

            ProjectCreated (Err err) ->
                let
                    errorMessages =
                        case err of
                            Http.BadStatus response ->
                                response.body
                                    |> decodeString (field "errors" errorsDecoder)
                                    |> Result.withDefault []

                            _ ->
                                [ ( "", "Unable to process project." ) ]
                in
                    { model
                        | submitting = False
                        , serverErrors = List.map serverErrorToFormError errorMessages
                    }
                        => Cmd.none

            ProjectCreated (Ok project) ->
                { model
                    | projects = project :: model.projects
                    , submitting = False
                    , formCollapsed = True
                    , form = initialForm
                }
                    => Cmd.none



-- VALIDATION --


type alias Error =
    ( Field, String )


validate : Validator ( Field, String ) ProjectForm
validate =
    Validate.all
        [ (.name >> .value) >> (ifBelowLength 3) (Name => "Name must be over 2 characters.")
        , (.name >> .value) >> (ifAboveLength 128) (Name => "Name must be less than 129 characters.")
        , (.repository >> .value) >> (ifBelowLength 8) (Repository => "Repository must be over 7 characters.")
        , (.repository >> .value) >> (ifAboveLength 128) (Repository => "Repository must less than 129 characters.")
        , (.privateKey >> .value) >> (ifBelowLength 8) (PrivateKey => "Private key must be over 7 characters.")
        ]


errorsDecoder : Decoder (List ( String, String ))
errorsDecoder =
    decode (\name repository privateKey -> List.concat [ name, repository, privateKey ])
        |> optionalError "name"
        |> optionalError "repository"
        |> optionalError "key"


optionalError : String -> Decoder (List ( String, String ) -> a) -> Decoder a
optionalError fieldName =
    let
        errorToTuple errorMessage =
            ( fieldName, errorMessage )
    in
        optional fieldName (Decode.list (Decode.map errorToTuple string)) []
