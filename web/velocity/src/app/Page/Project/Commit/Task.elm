module Page.Project.Commit.Task exposing (..)

-- EXTERNAL --

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Task exposing (Task)
import Dict exposing (Dict)
import Json.Encode as Encode
import Bootstrap.Modal as Modal
import Navigation
import Dom


-- INTERNAL --

import Context exposing (Context)
import Component.BuildLog as BuildLog
import Component.BuildForm as BuildForm
import Component.DropdownFilter as DropdownFilter
import Data.Device as Device
import Data.AuthToken as AuthToken exposing (AuthToken)
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStream as BuildStream exposing (Id, BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Util exposing ((=>))
import Views.Page as Page
import Page.Helpers exposing (formatDateTime, sortByDatetime)
import Request.Commit
import Request.Errors
import Route


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : BuildForm.Context
    , formModalVisibility : Modal.Visibility
    , selected : Maybe Build.Id
    , frame : Frame
    , buildDropdownState : DropdownFilter.DropdownState
    , buildFilterTerm : String
    }


type alias FromBuild =
    Build.Id


type alias ToBuild =
    Build.Id


type Stream
    = Stream BuildStream.Id


type Frame
    = BuildFrame BuildType
    | BlankFrame


type BuildType
    = LoadedBuild Build.Id BuildLog.Model
    | LoadingBuild (Maybe FromBuild) (Maybe ToBuild)


init : Context -> Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Task -> Maybe Build.Id -> List Build -> Task PageLoadError Model
init context session id hash task selected builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        init =
            maybeBuildToModel context task maybeAuthToken
    in
        case selected of
            Just id ->
                id
                    |> Build.findBuild builds
                    |> init

            Nothing ->
                builds
                    |> List.reverse
                    |> List.head
                    |> init


maybeBuildToModel :
    Context
    -> ProjectTask.Task
    -> Maybe AuthToken
    -> Maybe Build
    -> Task PageLoadError Model
maybeBuildToModel context task maybeAuthToken maybeBuild =
    let
        selected =
            Maybe.map .id maybeBuild

        init =
            initialModel task selected
    in
        case maybeBuild of
            Just b ->
                BuildLog.init context task maybeAuthToken b
                    |> Task.map (LoadedBuild b.id >> BuildFrame >> init)
                    |> Task.mapError handleLoadError

            Nothing ->
                Task.succeed (init BlankFrame)


initialModel : ProjectTask.Task -> Maybe Build.Id -> Frame -> Model
initialModel task selected frame =
    { task = task
    , toggledStep = Nothing
    , form = BuildForm.init task
    , formModalVisibility = Modal.hidden
    , selected = selected
    , frame = frame
    , buildDropdownState = DropdownFilter.initialDropdownState
    , buildFilterTerm = ""
    }


handleLoadError : a -> PageLoadError
handleLoadError _ =
    pageLoadError Page.Project "Project unavailable."



-- SUBSCRIPTIONS --


subscriptions : List Build -> Model -> Sub Msg
subscriptions builds model =
    let
        buildModal =
            Modal.subscriptions model.formModalVisibility AnimateFormModal

        buildDropdown =
            buildFilterContext model builds
                |> DropdownFilter.subscriptions buildDropdownFilterConfig

        buildOutput =
            case model.frame of
                BuildFrame (LoadedBuild _ buildOutputModel) ->
                    BuildLog.subscriptions buildOutputModel
                        |> Sub.map BuildLogMsg

                _ ->
                    Sub.none
    in
        Sub.batch [ buildModal, buildDropdown, buildOutput ]



-- CHANNELS --


events : Model -> Dict String (List ( String, Encode.Value -> Msg ))
events model =
    case model.frame of
        BuildFrame (LoadedBuild _ buildOutputModel) ->
            BuildLog.events buildOutputModel
                |> mapEvents BuildLogMsg

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe CommitRoute.Route -> List String
leaveChannels model route =
    let
        isTask task =
            model.task.name == task

        isBuild routeBuild =
            case ( routeBuild, model.selected ) of
                ( Just routeBuildId, Just selectedId ) ->
                    routeBuildId == selectedId

                _ ->
                    False

        channels =
            case model.frame of
                BuildFrame (LoadedBuild _ buildOutputModel) ->
                    BuildLog.leaveChannels buildOutputModel

                _ ->
                    []
    in
        case route of
            Just (CommitRoute.Task task maybeBuild) ->
                if not (isTask task) || not (isBuild maybeBuild) then
                    channels
                else
                    []

            _ ->
                channels


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)


buildDropdownFilterConfig : DropdownFilter.Config Msg Build
buildDropdownFilterConfig =
    { dropdownMsg = BuildFilterDropdownMsg
    , termMsg = BuildFilterTermMsg
    , noOpMsg = NoOp_
    , selectItemMsg = SelectBuild
    , labelFn = (.createdAt >> formatDateTime)
    , icon = (strong [] [ text "Build: " ])
    , showFilter = True
    , showAllItemsItem = False
    }


buildFilterContext : Model -> List Build -> DropdownFilter.Context Build
buildFilterContext { frame, buildDropdownState, buildFilterTerm, selected } builds =
    { items = sortByDatetime .createdAt builds
    , dropdownState = buildDropdownState
    , filterTerm = buildFilterTerm
    , selectedItem = Maybe.andThen (Build.findBuild builds) selected
    }



--
-- VIEW --


view : Device.Size -> Project -> Commit -> Model -> List Build -> Html Msg
view deviceSize project commit model builds =
    div []
        [ viewTabFrame model builds commit
        , viewFormModal model.task model.form model.formModalVisibility
        ]


viewHeader : Model -> List Build -> Html Msg
viewHeader model builds =
    div [ class "mb-4" ]
        [ viewTaskHeading model.task
        , viewToolbar model builds
        ]


viewTaskHeading : ProjectTask.Task -> Html Msg
viewTaskHeading task =
    h4 [] [ text (ProjectTask.nameToString task.name) ]


viewFormModal : ProjectTask.Task -> BuildForm.Context -> Modal.Visibility -> Html Msg
viewFormModal task form visibility =
    let
        hasFields =
            not (List.isEmpty form.fields)

        basicModal =
            Modal.config CloseFormModal
                |> Modal.withAnimation AnimateFormModal
                |> Modal.large
                |> Modal.hideOnBackdropClick True
                |> Modal.h5 []
                    [ text "Start "
                    , strong [] [ text <| ProjectTask.nameToString task.name ]
                    ]
                |> Modal.footer [] [ BuildForm.viewSubmitButton buildFormConfig form ]

        modal =
            if hasFields then
                Modal.body [] (BuildForm.view buildFormConfig form) basicModal
            else
                Modal.body [] [ viewNoParametersAlert ] basicModal
    in
        Modal.view visibility modal


viewNoParametersAlert : Html msg
viewNoParametersAlert =
    div [ class "alert alert-info m-0" ]
        [ i [ class "fa fa-info-circle" ] []
        , text " No parameters required"
        ]


viewToolbar : Model -> List Build -> Html Msg
viewToolbar model builds =
    let
        buildsDropdown =
            buildFilterContext model builds
                |> DropdownFilter.view buildDropdownFilterConfig

        shouldDisplayBuildsDropdown =
            List.length builds > 1

        newBuildButton =
            button
                [ class "btn btn-primary"
                , onClick OpenFormModal
                ]
                [ text
                    (if List.isEmpty builds then
                        "Run task"
                     else
                        "Run again"
                    )
                ]

        timeline =
            case model.frame of
                BuildFrame (LoadedBuild buildId buildOutputModel) ->
                    case findBuild builds buildId of
                        Just build ->
                            BuildLog.viewTimeline build buildOutputModel model.task
                                |> Html.map BuildLogMsg

                        Nothing ->
                            text ""

                _ ->
                    text ""
    in
        div [ class "d-flex" ]
            [ Util.viewIf shouldDisplayBuildsDropdown <| div [ class "pr-4" ] [ buildsDropdown ]
            , div [ class "flex-fill flex-grow-1 d-none d-sm-block" ] [ timeline ]
            , div [ class "pl-4" ] [ newBuildButton ]
            ]


findBuild : List Build -> Build.Id -> Maybe Build
findBuild builds id =
    builds
        |> List.filter (\a -> a.id == id)
        |> List.head


viewTabFrame : Model -> List Build -> Commit -> Html Msg
viewTabFrame model builds commit =
    if List.isEmpty builds then
        viewNoBuildsAlert model.task commit
    else
        case model.frame of
            BlankFrame ->
                text ""

            BuildFrame (LoadedBuild buildId buildOutputModel) ->
                case findBuild builds buildId of
                    Just build ->
                        BuildLog.view build model.task buildOutputModel
                            |> Html.map BuildLogMsg

                    Nothing ->
                        text ""

            BuildFrame (LoadingBuild _ _) ->
                viewToolbar model builds


viewNoBuildsAlert : ProjectTask.Task -> Commit -> Html Msg
viewNoBuildsAlert task commit =
    let
        icon =
            i [ class "fa fa-info-circle" ] []

        preTaskNameText =
            text " Task "

        taskNameText =
            strong [] [ text (ProjectTask.nameToString task.name) ]

        postTaskNameText =
            text " has not run yet for commit "

        commitShaText =
            strong [] [ text (Commit.truncateHash commit.hash) ]
    in
        div [ class "alert alert-info mt-4" ]
            [ icon
            , preTaskNameText
            , taskNameText
            , postTaskNameText
            , commitShaText
            ]


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route.Route, String )
breadcrumb project commit task =
    [ ( CommitRoute.Task task.name Nothing |> ProjectRoute.Commit commit.hash |> Route.Project project.slug
      , ProjectTask.nameToString task.name
      )
    ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput BuildForm.InputFormField String
    | OnChange BuildForm.ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Request.Errors.HttpError Build)
    | SelectBuild (Maybe Build)
    | BuildLoaded (Result Request.Errors.HttpError (Maybe BuildType))
    | BuildLogMsg BuildLog.Msg
    | CloseFormModal
    | AnimateFormModal Modal.Visibility
    | OpenFormModal
    | BuildFilterDropdownMsg DropdownFilter.DropdownState
    | BuildFilterTermMsg String
    | NoOp_


type ExternalMsg
    = NoOp
    | AddBuild Build
    | UpdateBuild Build


buildFormConfig : BuildForm.Config Msg
buildFormConfig =
    { submitMsg = SubmitForm
    , onChangeMsg = OnChange
    , onInputMsg = OnInput
    }


update : Context -> Project -> Commit -> List Build -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context project commit builds session msg model =
    let
        projectSlug =
            project.slug

        commitHash =
            commit.hash

        taskName =
            model.task.name

        maybeAuthToken =
            Maybe.map .token session.user
    in
        case msg of
            OpenFormModal ->
                let
                    form =
                        BuildForm.init model.task

                    focusFirstField =
                        form
                            |> BuildForm.firstId
                            |> Maybe.map Dom.focus
                            |> Maybe.map (Task.attempt (always NoOp_))
                            |> Maybe.withDefault Cmd.none
                in
                    { model
                        | formModalVisibility = Modal.shown
                        , form = form
                    }
                        => focusFirstField
                        => NoOp

            CloseFormModal ->
                { model
                    | formModalVisibility = Modal.hidden
                    , form = BuildForm.init model.task
                }
                    => Cmd.none
                    => NoOp

            AnimateFormModal visibility ->
                { model | formModalVisibility = visibility }
                    => Cmd.none
                    => NoOp

            ToggleStep maybeStep ->
                { model | toggledStep = maybeStep }
                    => Cmd.none
                    => NoOp

            OnInput field value ->
                { model | form = BuildForm.updateInput field value model.form }
                    => Cmd.none
                    => NoOp

            OnChange field maybeIndex ->
                { model | form = BuildForm.updateSelect field maybeIndex model.form }
                    => Cmd.none
                    => NoOp

            SubmitForm ->
                let
                    cmdFromAuth authToken =
                        authToken
                            |> Request.Commit.createBuild context projectSlug commitHash taskName (BuildForm.submitParams model.form)
                            |> Task.attempt BuildCreated

                    cmd =
                        session
                            |> Session.attempt "create build" cmdFromAuth
                            |> Tuple.second
                in
                    model
                        => cmd
                        => NoOp

            BuildLoaded (Ok (Just loadedBuild)) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLoaded _ ->
                model
                    => Cmd.none
                    => NoOp

            BuildCreated (Ok build) ->
                let
                    route =
                        CommitRoute.Task model.task.name (Just build.id)
                            |> ProjectRoute.Commit commit.hash
                            |> Route.Project project.slug
                in
                    model
                        => Navigation.newUrl (Route.routeToString route)
                        => AddBuild build

            BuildCreated (Err _) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLogMsg subMsg ->
                case model.frame of
                    BuildFrame (LoadedBuild id outputModel) ->
                        let
                            ( newOutputModel, newOutputCmd ) =
                                BuildLog.update subMsg outputModel
                        in
                            { model | frame = BuildFrame (LoadedBuild id newOutputModel) }
                                => Cmd.map BuildLogMsg newOutputCmd
                                => NoOp

                    _ ->
                        model
                            => Cmd.none
                            => NoOp

            BuildFilterDropdownMsg state ->
                { model | buildDropdownState = state }
                    => Cmd.none
                    => NoOp

            BuildFilterTermMsg term ->
                { model | buildFilterTerm = term }
                    => Cmd.none
                    => NoOp

            SelectBuild maybeBuild ->
                let
                    fromBuild =
                        case model.frame of
                            BuildFrame (LoadedBuild id _) ->
                                Just id

                            _ ->
                                Nothing

                    toBuild =
                        Maybe.map .id maybeBuild

                    route =
                        toBuild
                            |> CommitRoute.Task model.task.name
                            |> ProjectRoute.Commit commitHash
                            |> Route.Project project.slug
                in
                    { model
                        | frame = BuildFrame (LoadingBuild fromBuild toBuild)
                        , buildDropdownState = DropdownFilter.initialDropdownState
                        , selected = toBuild
                    }
                        => Route.modifyUrl route
                        => NoOp

            NoOp_ ->
                model
                    => Cmd.none
                    => NoOp
