- Initialize go module
    ```
    go mod init github.com/benauro/kube-cdn
    ```

- Initialize kubebuilder project
    ```
    kubebuilder init --domain benauro.gg --repo github.com/benauro/kube-cdn
    ```

- Create API (v3)
    ```
    kubebuilder create api --group cdn --version v3 --kind ContentDeliveryNetwork
    ```
    ```
    make manifests
    ```

- Modify Custom Resource Definition (CRD)

    [Content Delivery Network Types (Go file)](../api/v3/contentdeliverynetwork_types.go)

- Modify Controller

    [Content Delivery Network Controller (Go file)](../internal/controller/contentdeliverynetwork_controller.go)

- Run Controller
    ```
    make run
    ```