using UnityEngine;

public class FrustrumVisualizer : MonoBehaviour
{

    public float fov = 60.0f;
    public float maxRange = 10.0f;
    public float minRange = 0.1f;
    public float aspect = 16.0f / 9.0f;

    void OnDrawGizmosSelected()
    {
        Gizmos.matrix = transform.localToWorldMatrix;
        Gizmos.color = Color.yellow;
        Gizmos.DrawFrustum(Vector3.zero, fov, maxRange, minRange, aspect);
    } 
}
