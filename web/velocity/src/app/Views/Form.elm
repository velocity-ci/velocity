module Views.Form exposing (input, password, select, textarea, viewErrors, viewSpinner)

import Html exposing (Attribute, Html, div, fieldset, i, label, li, small, span, text, ul)
import Html.Attributes exposing (class, for, id, type_)


password :
    { r
        | name : String
        , label : String
        , help : Maybe String
        , errors : List ( field, String )
    }
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
password { name, label, help, errors } attrs =
    control name label errors help Html.input ([ type_ "password" ] ++ attrs)


input :
    { r
        | name : String
        , label : String
        , help : Maybe String
        , errors : List ( field, String )
    }
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
input { name, label, help, errors } attrs =
    control name label errors help Html.input ([ type_ "text" ] ++ attrs)


select :
    { r
        | name : String
        , label : String
        , help : Maybe String
        , errors : List ( field, String )
    }
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
select { name, label, help, errors } attrs =
    control name label errors help Html.select ([] ++ attrs)


textarea :
    { r
        | name : String
        , label : String
        , help : Maybe String
        , errors : List ( field, String )
    }
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
textarea { name, label, help, errors } attrs =
    control name label errors help Html.textarea attrs


viewErrors : List ( a, String ) -> Html msg
viewErrors errors =
    errors
        |> List.map (\( _, error ) -> li [] [ text error ])
        |> ul [ class "error-messages" ]


viewSpinner : Html msg
viewSpinner =
    span []
        [ i [ class "fa fa-circle-o-notch fa-spin fa-fw" ] []
        , span [ class "sr-only" ] [ text "Loading..." ]
        ]



-- INTERNAL --


viewErrorFeedback : ( field, String ) -> Html msg
viewErrorFeedback error =
    div [ class "invalid-feedback" ] [ text (Tuple.second error) ]


control :
    String
    -> String
    -> List ( field, String )
    -> Maybe String
    -> (List (Attribute msg) -> List (Html msg) -> Html msg)
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
control name labelText errors maybeHelp element attributes children =
    let
        help =
            case maybeHelp of
                Just helpText ->
                    small [ class "form-text text-muted" ] [ text helpText ]

                Nothing ->
                    text ""
    in
        div [ class "form-group" ]
            ([ label [ for name ] [ text labelText ]
             , element (class "form-control" :: id name :: attributes) children
             , help
             ]
                ++ List.map viewErrorFeedback errors
            )
