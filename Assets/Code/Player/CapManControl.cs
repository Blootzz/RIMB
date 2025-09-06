using PlasticGui.WorkspaceWindow;
using Unity.Netcode;
using UnityEditor.Animations;
using UnityEngine;

public class CapManControl : NetworkBehaviour
{

    [SerializeField]
    private GameObject cameraTransformTarget;
    private Matrix4x4 origCamTranform;
    private NetworkVariable<Color> playerColor = new(Color.black, NetworkVariableReadPermission.Everyone, NetworkVariableWritePermission.Owner);

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

    public override void OnNetworkSpawn()
    {
        if (IsOwner)
        {
            GameObject camera = GameObject.FindWithTag("MainCamera");
            origCamTranform = camera.transform.localToWorldMatrix;
            camera.transform.SetParent(cameraTransformTarget.transform);
            camera.transform.SetLocalPositionAndRotation(Vector3.zero, Quaternion.identity);
        }
        else
        {
            gameObject.layer = 0;
            foreach (Transform child in gameObject.transform.GetComponentsInChildren<Transform>()) {
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

}
