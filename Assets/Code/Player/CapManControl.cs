using UnityEngine;
using Unity.Netcode;
using UnityEngine.InputSystem;

public class CapManControl : NetworkBehaviour
{

    [Range(0.01f, 2.0f)]
    public float lookSensitivity = 0.1f;
    public float moveSpeed = 10.0f;

    private InputAction move = null;
    private InputAction look = null;

    [SerializeField]
    private GameObject cameraTransformTarget;
    private Matrix4x4 origCamTranform;
    private NetworkVariable<Color> playerColor = new(
        Color.black,
        NetworkVariableReadPermission.Everyone,
        NetworkVariableWritePermission.Owner); // NVs are global, so this will need to be changed later.

    public Color PlayerColor
    {
        get
        {
            return playerColor.Value;
        }
        set
        {
            if (IsOwner)
                playerColor.Value = value; // Enforce the write constraints
        }
    }

    public float LookSensitivity
    {
        get
        {
            return lookSensitivity;
        }
        set
        {
            if (IsOwner)
            {
                lookSensitivity = value;
            }
        }
    }

    public override void OnNetworkSpawn()
    {
        if (IsOwner)
        {
            GameObject camera = GameObject.FindWithTag("MainCamera");
            origCamTranform = camera.transform.localToWorldMatrix;
            camera.transform.SetParent(cameraTransformTarget.transform);
            camera.transform.SetLocalPositionAndRotation(Vector3.zero, Quaternion.identity);
            move = InputSystem.actions.FindAction("Move");
            look = InputSystem.actions.FindAction("Look");
            Cursor.lockState = CursorLockMode.Locked;
        }
        else
        {
            gameObject.layer = 0;
            foreach (Transform child in gameObject.transform.GetComponentsInChildren<Transform>())
            {
                child.gameObject.layer = 0;
            }
        }
        playerColor.OnValueChanged += OnColorChange;
    }

    public override void OnNetworkDespawn()
    {
        if (IsOwner)
        {
            GameObject camera = GameObject.FindWithTag("MainCamera");
            camera.transform.SetParent(null);
            camera.transform.SetLocalPositionAndRotation(origCamTranform.GetPosition(), origCamTranform.rotation);
            Cursor.lockState = CursorLockMode.None;
        }
        playerColor.OnValueChanged -= OnColorChange;
        base.OnNetworkDespawn();
    }

    private void OnColorChange(Color oldValue, Color newValue)
    {
        Renderer renderer = gameObject.GetComponent<Renderer>();
        if (renderer == null)
            return;
        Material material = renderer.material;
        if (material == null)
            return;
        material.SetColor("_BaseColor", newValue);
    }

    private void ApplyMovementToPlayer(Vector2 direction, float magnitude)
    {
        Rigidbody player = gameObject.GetComponent<Rigidbody>();
        Vector3 currentVelocity = player.linearVelocity;
        Vector3 vel = new(direction.x, 0, direction.y);
        vel = (transform.rotation * vel).normalized * magnitude;
        Vector3 targetVelocity = new Vector3(vel.x, currentVelocity.y, vel.z);
        Vector3 velocityDiff = targetVelocity - currentVelocity;
        Vector3 force = velocityDiff * player.mass / Time.fixedDeltaTime;
        float maxAccel = 30f;
        force = Vector3.ClampMagnitude(force, maxAccel * player.mass);
        player.AddForce(force);
    }

    private void ApplyLookToPlayer(Vector2 direction, float sensitivityMultiplier = 1.0f)
    {
        gameObject.transform.Rotate(new Vector3(0, direction.x * sensitivityMultiplier, 0), Space.Self);
        Vector3 locRot = cameraTransformTarget.transform.localEulerAngles;
        float pitch = Mathf.DeltaAngle(0f, locRot.x);
        locRot.x = Mathf.Clamp(pitch - direction.y * sensitivityMultiplier, -80, 80);
        cameraTransformTarget.transform.localEulerAngles = locRot;
    }

    void FixedUpdate()
    {
        // Physics related stuff MUST happen here bc the physics engine runs on the same fixed update.
        if (Cursor.lockState != CursorLockMode.None)
        {
            Vector2 moveVec = (Vector2)(move?.ReadValue<Vector2>());
            if (moveVec != null)
                ApplyMovementToPlayer(moveVec, moveSpeed);
        }
    }

    void Update()
    {
        if (Cursor.lockState != CursorLockMode.None)
        {
            Vector2 lookVec = (Vector2)(look?.ReadValue<Vector2>());
            if (lookVec != null && lookVec != Vector2.zero)
                ApplyLookToPlayer(lookVec, lookSensitivity);
        }
    }

}
