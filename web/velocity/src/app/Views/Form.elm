module Views.Form exposing (viewErrors, input, textarea, password, viewSpinner)

import Html exposing (fieldset, ul, li, Html, Attribute, text, div, label, span, i)
import Html.Attributes exposing (class, type_, for, id)


password :
    String
    -> String
    -> List ( field, String )
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
password name labelText errors attrs =
    control name labelText errors Html.input ([ type_ "password" ] ++ attrs)


input :
    String
    -> String
    -> List ( field, String )
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
input name labelText errors attrs =
    control name labelText errors Html.input ([ type_ "text" ] ++ attrs)


textarea :
    String
    -> String
    -> List ( field, String )
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
textarea name labelText errors attrs =
    control name labelText errors Html.textarea attrs


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
    -> (List (Attribute msg) -> List (Html msg) -> Html msg)
    -> List (Attribute msg)
    -> List (Html msg)
    -> Html msg
control name labelText errors element attributes children =
    div [ class "form-group" ]
        ([ label [ for name ] [ text labelText ]
         , element (class "form-control" :: id name :: attributes) children
         ]
            ++ List.map viewErrorFeedback errors
        )
